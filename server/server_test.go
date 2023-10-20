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
					go socket.ProcessMerge(func(pkt *packet.Packet) error {
						fmt.Println(len(pkt.ToBytes()))
						return nil
					})
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

func TestSocket(t *testing.T) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second})
	if err := sock.Open(); err == nil {
		for i := 0; i < 1<<10; i++ {
			fmt.Println(i)
			sock.WriteWithLen([]byte(fmt.Sprint(i)))
		}
	}
	Process(sock, process)
}

func TestSocketMerge(t *testing.T) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, SnappyMergeData: true})
	if err := sock.Open(); err == nil {
		for i := 0; i < 100; i++ {
			go sock.WriteWithMerge([]byte(fmt.Sprint(i)))
		}
	}
	sock.ProcessMerge(func(pkt *packet.Packet) (err error) {
		time.Sleep(3 * time.Second)
		sock.WriteWithMerge(pkt.ToBytes())
		return
	})
}

func BenchmarkSocketMerge(b *testing.B) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, SnappyMergeData: true})
	if sock.Open() != nil {
		panic("open failed")
	}
	i := 0
	go sock.ProcessMerge(func(pkt *packet.Packet) (err error) {
		// fmt.Println(string(pkt.ToBytes()))
		return
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			sock.WriteWithMerge([]byte(fmt.Sprint(i, "123456789")))
		}
	})
}
