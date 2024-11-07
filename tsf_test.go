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
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func Test_tsf(t *testing.T) {
	cfg := &TsfConfig{ListenAddr: ":20001", TConfiguration: &TConfiguration{ProcessMerge: true}}
	tx := &TContext{}
	tx.OnOpen = func(socket TsfSocket) error {
		fmt.Println("OnOpen:", socket.ID())
		return nil
	}
	tx.Handler = func(socket TsfSocket, packet *Packet) error {
		fmt.Println("server recv:", string(packet.ToBytes()))
		if _, err := socket.WriteWithMerge(packet.ToBytes()); err != nil {
			panic(err)
		}
		return nil
	}
	tx.OnClose = func(socket TsfSocket) error {
		fmt.Println("OnClose:", socket.ID())
		return nil
	}
	tx.OnError = func(err error, socket TsfSocket) error {
		fmt.Println("OnError:", socket.ID(), ",err:", err)
		return nil
	}
	s, err := NewTsf(cfg, tx)
	if err != nil {
		t.Fatal(err)
	}
	s.Serve()
}

func Test_tsfclient(t *testing.T) {
	tx := &TContext{}
	var i atomic.Int32
	tx.OnOpen = func(socket TsfSocket) error {
		fmt.Println("OnOpen:", socket.ID())
		if _, err := socket.WriteWithMerge([]byte(fmt.Sprint(i.Add(1)))); err != nil {
			panic(err)
		}
		return nil
	}
	tx.Handler = func(socket TsfSocket, packet *Packet) error {
		fmt.Println("client recv:", string(packet.ToBytes()))
		if i.Load() < 10 {
			for range 20 {
				if _, err := socket.WriteWithMerge([]byte(fmt.Sprint(i.Add(1)))); err != nil {
					panic(err)
				}
			}

		}
		return nil
	}
	tx.OnClose = func(socket TsfSocket) error {
		fmt.Println("OnClose:", socket.ID())
		return nil
	}
	tx.OnError = func(err error, socket TsfSocket) error {
		fmt.Println("OnError:", socket.ID(), ",err:", err)
		return nil
	}

	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, ProcessMerge: true})

	if err := sock.Open(); err == nil {
		sock.On(tx)
	} else {
		panic(err.Error())
	}
}
