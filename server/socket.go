/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package server

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/tsf/util"
)

type TSocket struct {
	conn      *socketConn
	addr      net.Addr
	cfg       *TConfiguration
	_dataChan chan []byte
	mux       *sync.Mutex
	_incount  int64
}

// tcpAddr is a naive implementation of net.Addr that does nothing extra.
type tcpAddr string

var _ net.Addr = tcpAddr("")

func (ta tcpAddr) Network() string {
	return "tcp"
}

func (ta tcpAddr) String() string {
	return string(ta)
}

// Deprecated: Use NewTSocketConf instead.
// func NewTSocket(hostPort string) (*TSocket, error) {
// 	return NewTSocketConf(hostPort, &TConfiguration{
// 		noPropagation: true,
// 	}), nil
// }

// NewTSocketConf creates a net.Conn-backed TTransport, given a host and port.
//
// Example:
//
//	trans, err := thrift.NewTSocketConf("localhost:9090", &TConfiguration{
//	    ConnectTimeout: time.Second, // Use 0 for no timeout
//	    SocketTimeout:  time.Second, // Use 0 for no timeout
//	})
func NewTSocketConf(hostPort string, conf *TConfiguration) *TSocket {
	return NewTSocketFromAddrConf(tcpAddr(hostPort), conf)
}

// Deprecated: Use NewTSocketConf instead.
// func NewTSocketTimeout(hostPort string, connTimeout time.Duration, soTimeout time.Duration) (*TSocket, error) {
// 	return NewTSocketConf(hostPort, &TConfiguration{
// 		ConnectTimeout: connTimeout,
// 		SocketTimeout:  soTimeout,

// 		noPropagation: true,
// 	}), nil
// }

// NewTSocketFromAddrConf creates a TSocket from a net.Addr
func NewTSocketFromAddrConf(addr net.Addr, conf *TConfiguration) *TSocket {
	return &TSocket{
		addr: addr,
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

// Deprecated: Use NewTSocketFromAddrConf instead.
// func NewTSocketFromAddrTimeout(addr net.Addr, connTimeout time.Duration, soTimeout time.Duration) *TSocket {
// 	return NewTSocketFromAddrConf(addr, &TConfiguration{
// 		ConnectTimeout: connTimeout,
// 		SocketTimeout:  soTimeout,

// 		noPropagation: true,
// 	})
// }

// NewTSocketFromConnConf creates a TSocket from an existing net.Conn.
func NewTSocketFromConnConf(conn net.Conn, conf *TConfiguration) *TSocket {
	return &TSocket{
		conn: wrapSocketConn(conn),
		addr: conn.RemoteAddr(),
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

// Deprecated: Use NewTSocketFromConnConf instead.
func NewTSocketFromConnTimeout(conn net.Conn, socketTimeout time.Duration) *TSocket {
	return NewTSocketFromConnConf(conn, &TConfiguration{
		SocketTimeout: socketTimeout,

		// noPropagation: true,
	})
}

// SetTConfiguration implements TConfigurationSetter.
//
// It can be used to set connect and socket timeouts.
func (p *TSocket) SetTConfiguration(conf *TConfiguration) {
	p.cfg = conf
}

// Sets the connect timeout
func (p *TSocket) SetConnTimeout(timeout time.Duration) error {
	if p.cfg == nil {
		p.cfg = &TConfiguration{
			// noPropagation: true,
		}
	}
	p.cfg.ConnectTimeout = timeout
	return nil
}

// Sets the socket timeout
func (p *TSocket) SetSocketTimeout(timeout time.Duration) error {
	if p.cfg == nil {
		p.cfg = &TConfiguration{
			// noPropagation: true,
		}
	}
	p.cfg.SocketTimeout = timeout
	return nil
}

func (p *TSocket) pushDeadline(read, write bool) {
	var t time.Time
	if timeout := p.cfg.GetSocketTimeout(); timeout > 0 {
		t = time.Now().Add(time.Duration(timeout))
	}
	if read && write {
		p.conn.SetDeadline(t)
	} else if read {
		p.conn.SetReadDeadline(t)
	} else if write {
		p.conn.SetWriteDeadline(t)
	}
}

// Connects the socket, creating a new socket object if necessary.
func (p *TSocket) Open() error {
	if p.conn.isValid() {
		return NewTTransportException(ALREADY_OPEN, "Socket already connected.")
	}
	if p.addr == nil {
		return NewTTransportException(NOT_OPEN, "Cannot open nil address.")
	}
	if len(p.addr.Network()) == 0 {
		return NewTTransportException(NOT_OPEN, "Cannot open bad network name.")
	}
	if len(p.addr.String()) == 0 {
		return NewTTransportException(NOT_OPEN, "Cannot open bad address.")
	}
	var err error
	if p.conn, err = createSocketConnFromReturn(net.DialTimeout(
		p.addr.Network(),
		p.addr.String(),
		p.cfg.GetConnectTimeout(),
	)); err != nil {
		return &tTransportException{
			typeId: NOT_OPEN,
			err:    err,
			msg:    err.Error(),
		}
	}
	p.addr = p.conn.RemoteAddr()
	return nil
}

// Retrieve the underlying net.Conn
func (p *TSocket) Conn() net.Conn {
	return p.conn
}

// Returns true if the connection is open
func (p *TSocket) IsOpen() bool {
	return p.conn.IsOpen()
}

// Closes the socket.
func (p *TSocket) Close() error {
	return p.conn.Close()
}

// Returns the remote address of the socket.
func (p *TSocket) Addr() net.Addr {
	return p.addr
}

func (p *TSocket) Read(buf []byte) (int, error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	p.pushDeadline(true, false)
	// NOTE: Calling any of p.IsOpen, p.conn.read0, or p.conn.IsOpen between
	// p.pushDeadline and p.conn.Read could cause the deadline set inside
	// p.pushDeadline being reset, thus need to be avoided.
	n, err := p.conn.Read(buf)
	return n, NewTTransportExceptionFromError(err)
}

func (p *TSocket) Write(buf []byte) (int, error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	p.pushDeadline(false, true)
	return p.conn.Write(buf)
}

func (p *TSocket) WriteWithLen(buf []byte) (int, error) {
	ln := 4
	if p.cfg.Packet64Bits {
		ln = 8
	}
	bys := make([]byte, ln+len(buf))
	if p.cfg.Packet64Bits {
		copy(bys[:ln], util.Int64ToBytes(int64(len(buf))))
	} else {
		copy(bys[:ln], util.Int32ToBytes(int32(len(buf))))
	}
	copy(bys[ln:], buf)
	return p.Write(bys)
}

func (p *TSocket) WriteWithMerge(buf []byte) {
	if p._dataChan == nil {
		p.mux.Lock()
		if p._dataChan == nil {
			p._dataChan = make(chan []byte, 1<<13)
		}
		p.mux.Unlock()
	}
	p._dataChan <- buf
	atomic.AddInt64(&p._incount, 1)
	if p.mux.TryLock() {
		go writeMerge(p)
	}
}

func (p *TSocket) Flush(ctx context.Context) error {
	return nil
}

func (p *TSocket) Interrupt() error {
	if !p.conn.isValid() {
		return nil
	}
	return p.conn.Close()
}

func (p *TSocket) RemainingBytes() (num_bytes uint64) {
	const maxSize = ^uint64(0)
	return maxSize // the truth is, we just don't know unless framed is used
}
