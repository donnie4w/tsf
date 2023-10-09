/**
 * Copyright 2023 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/donnie4w/tsf/packet"
)

func TestServer(t *testing.T) {
	if serversocket, err := NewTServerSocketTimeout(":20001", 10*time.Second); err == nil {
		if err = serversocket.Listen(); err == nil {
			for {
				if socket, err := serversocket.Accept(); err == nil {
					go Process(socket, process)
				}
			}
		}
	}
}

func TestServerMerge(t *testing.T) {
	if serversocket, err := NewTServerSocketTimeout(":20001", 10*time.Second); err == nil {
		if err = serversocket.Listen(); err == nil {
			for {
				if socket, err := serversocket.Accept(); err == nil {
					socket.SetTConfiguration(&TConfiguration{SnappyMergeData: true})
					go ProcessMerge(socket, process2)
				}
			}
		}
	}
}

func process(sock TsfSocket, pkt *packet.Packet) (err error) {
	fmt.Println(string(pkt.ToBytes()))
	time.Sleep(3 * time.Second)
	sock.WriteWithLen(pkt.ToBytes())
	return
}

func process2(sock TsfSocket, pkt *packet.Packet) (err error) {
	fmt.Println(string(pkt.ToBytes()))
	time.Sleep(3 * time.Second)
	sock.WriteWithMerge(pkt.ToBytes())
	return
}

func TestSocket(t *testing.T) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second})
	if err := sock.Open(); err == nil {
		bs := []byte("haha111")
		sock.WriteWithLen(bs)
	}
	Process(sock, process)
}

func TestSocketMerge(t *testing.T) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, SnappyMergeData: true})
	if err := sock.Open(); err == nil {
		bs := []byte("haha111")
		for i := 0; i < 10; i++ {
			sock.WriteWithMerge(append(bs, []byte(fmt.Sprint(i))...))
		}
	}
	ProcessMerge(sock, process2)
}
