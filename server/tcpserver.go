/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/go-logger/logger"
	. "github.com/donnie4w/tsf/conf"
	. "github.com/donnie4w/tsf/hashtable"
	. "github.com/donnie4w/tsf/packet"
	. "github.com/donnie4w/tsf/utils"
)

var idmap *HashTable = NewHashTable()
var seqIdmap *HashTable = NewHashTable()

func tcpStart() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	service := fmt.Sprint(":", CF.TcpPort)
	logger.Debug("tcp port:", CF.TcpPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		logger.Debug(conn.RemoteAddr().String(), " is conneted")
		if err == nil {
			go newConnction(conn)
		}
	}
}

func newConnction(conn net.Conn) {
	connection := &Connection{conn: conn, rmu: new(sync.Mutex), wmu: new(sync.Mutex), c: make(chan int, 0)}
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			connection.err(fmt.Sprint(err))
		}
		connection.close()
	}()
	go func() {
		catch()
		i := 0
		for !connection.closed {
			time.Sleep(1 * time.Second)
			if i > 30 {
				if connection.i > 3 || connection.ping() != nil {
					connection.close()
				} else {
					atomic.AddInt32(&connection.i, 1)
				}
				i = 0
			}
			i++
		}
	}()
	for {
		p, err := connection.readPacket()
		if err != nil {
			logger.Error("err.", err.Error())
			panic(fmt.Sprint(err))
		}
		err = connection.handler(p)
		if err != nil {
			logger.Error("err.", err.Error())
			panic(fmt.Sprint(err))
		}
	}
}

type Connection struct {
	conn   net.Conn
	rmu    *sync.Mutex
	wmu    *sync.Mutex
	closed bool
	c      chan int
	i      int32
}

func (this *Connection) close() {
	defer catch()
	this.conn.Close()
	this.closed = true
}

func (this *Connection) readPacket() (p *Packet, e error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			e = errors.New(fmt.Sprint(err))
		}
	}()
	headBs := make([]byte, 4)
	n, err := this.conn.Read(headBs)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errors.New("readPacket err")
	}
	length := int(BytesToInt(headBs))
	buf := bytes.NewBuffer([]byte{})
	i := 0
	for {
		bs := make([]byte, length-buf.Len())
		i, err = this.conn.Read(bs)
		if err == nil {
			buf.Write(bs[0:i])
		} else {
			panic(err.Error())
		}
		if length == buf.Len() {
			break
		}
	}
	bs := buf.Bytes()
	p = Wrap(bs)
	return
}

func (this *Connection) handler(p *Packet) (err error) {
	switch p.PType {
	case Auth: //auth
		if this.auth(p) == nil {
			p.Body = []byte{0}
			err = this.ack(p, Auth)
		} else {
			panic("no auth")
		}
	case Service: //service id
		this.service(p)
	case Ping: //ping
		p.Body = []byte{0}
		err = this.ack(p, AckPing)
	case Register: //register body
		err = this.register(p)
	case AckService: //ServiceId ack
	case AckRegister: //Register ack
		err = this.ackRegister(p)
	case AckPing: //Ping ack
		atomic.StoreInt32(&this.i, 0)
	default:
		err = errors.New(fmt.Sprint("Packet type err. ", p.PType))
	}
	return
}

func (this *Connection) ackRegister(p *Packet) (err error) {
	seqid := p.SeqId
	if seqid > 0 {
		conn := seqIdmap.Get(seqid).(*Connection)
		bs := p.GetPacket()
		_, err = conn.conn.Write(bs)
		seqIdmap.Del(seqid)
	}
	return
}

func (this *Connection) auth(p *Packet) (err error) {
	fmt.Println("auth:", string(p.Body))
	if CF.Auth != string(p.Body) {
		err = errors.New(fmt.Sprint("no auth:", string(p.Body)))
	}
	return
}

func (this *Connection) service(p *Packet) {
	fmt.Println(string(p.Serviceid))
	idmap.Put(string(p.Serviceid), this)
	return
}

func (this *Connection) register(p *Packet) (err error) {
	seqid := p.SeqId
	seqIdmap.Put(seqid, this)
	serviceid := p.Serviceid
	i := idmap.Get(string(serviceid))
	if i != nil {
		conn := idmap.Get(string(serviceid)).(*Connection)
		conn.conn.Write(p.GetPacket())
	} else {
		err = errors.New(fmt.Sprint("no serviceid:", serviceid))
	}
	return
}

func (this *Connection) write(p *Packet) (err error) {
	this.wmu.Lock()
	defer this.wmu.Unlock()
	_, err = this.conn.Write(p.GetPacket())
	return
}

func (this *Connection) ping() (err error) {
	p := NewPacket()
	p.PType = Ping
	err = this.write(p)
	return
}

func (this *Connection) ack(p *Packet, ack PacketType) (err error) {
	p.PType = ack
	err = this.write(p)
	return
}

func (this *Connection) err(s string) (err error) {
	defer catch()
	logger.Debug("err:", s)
	p := new(Packet)
	p.PType = Err
	p.Serviceid = []byte{0}
	p.Body = []byte(s)
	err = this.write(p)
	return
}

func catch() {
	if err := recover(); err != nil {
		logger.Error(err)
		logger.Error(string(debug.Stack()))
	}
}

func checkError(err error) {
	if err != nil {
		logger.Error(err.Error())
		panic(err.Error())
	}
}
