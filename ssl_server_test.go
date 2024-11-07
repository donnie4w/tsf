/*
 * Copyright (c) 2017 donnie4w <donnie4w@gmail.com>. All rights reserved.
 * Original source: https://github.com/donnie4w/tsf
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tsf

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"
)

func testSSLConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		panic(err)
	}
	// Initialize the TLS configuration.
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return cfg
}

func TestSSLServer(t *testing.T) {
	if server, err := NewTSSLServerSocketTimeout(":20001", testSSLConfig(), 10*time.Second); err == nil {
		if err = server.Listen(); err == nil {
			for {
				if socket, err := server.Accept(); err == nil {
					fmt.Println("socket accepted")
					go func() {
						err := Process(socket, processSSL)
						if err != nil {
							fmt.Println(err)
						}
					}()
				}
			}
		}
	}
}

func TestSSLServerMerge(t *testing.T) {
	if serversocket, err := NewTServerSocketTimeout(":20001", 10*time.Second); err == nil {
		if err = serversocket.Listen(); err == nil {
			for {
				if socket, err := serversocket.Accept(); err == nil {
					socket.SetTConfiguration(&TConfiguration{Snappy: true})
					go socket.ProcessMerge(func(pkt *Packet) error {
						fmt.Println(len(pkt.ToBytes()))
						return nil
					})
				}
			}
		}
	}
}

func processSSL(sock TsfSocket, pkt *Packet) (err error) {
	fmt.Println(string(pkt.ToBytes()))
	time.Sleep(3 * time.Second)
	sock.Write(pkt.ToBytes())
	return
}

func TestSSLSocket(t *testing.T) {
	sock := NewTSSLSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, TLSConfig: &tls.Config{InsecureSkipVerify: true}})
	if err := sock.Open(); err == nil {
		for i := 0; i < 1<<10; i++ {
			fmt.Println(i)
			if _, err := sock.Write([]byte(fmt.Sprint(i))); err != nil {
				panic(err)
			}
		}
	}
	//Process(sock, processSSL)
}

func TestSSLSocketMerge(t *testing.T) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, Snappy: true})
	if err := sock.Open(); err == nil {
		for i := 0; i < 100; i++ {
			go sock.WriteWithMerge([]byte(fmt.Sprint(i)))
		}
	}
	sock.ProcessMerge(func(pkt *Packet) (err error) {
		time.Sleep(3 * time.Second)
		sock.WriteWithMerge(pkt.ToBytes())
		return
	})
}

func BenchmarkSSLSocketMerge(b *testing.B) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, Snappy: true})
	if sock.Open() != nil {
		panic("open failed")
	}
	i := 0
	go sock.ProcessMerge(func(pkt *Packet) (err error) {
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
