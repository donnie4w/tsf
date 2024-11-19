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

import (
	"errors"
	"github.com/donnie4w/gofer/uuid"
	"net"
	"sync"
	"time"
)

type TServerSocket struct {
	listener      net.Listener
	addr          net.Addr
	clientTimeout time.Duration
	mu            sync.RWMutex
	interrupted   bool
}

func NewTServerSocket(listenAddr string) (*TServerSocket, error) {
	return NewTServerSocketTimeout(listenAddr, 0)
}

func NewTServerSocketTimeout(listenAddr string, clientTimeout time.Duration) (*TServerSocket, error) {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &TServerSocket{addr: addr, clientTimeout: clientTimeout}, nil
}

func (p *TServerSocket) Listen() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsListening() {
		return nil
	}
	l, err := net.Listen(p.addr.Network(), p.addr.String())
	if err != nil {
		return err
	}
	p.listener = l
	return nil
}

func (p *TServerSocket) Accept() (*TSocket, error) {
	p.mu.RLock()
	interrupted := p.interrupted
	p.mu.RUnlock()

	if interrupted {
		return nil, errTransportInterrupted
	}

	p.mu.Lock()
	listener := p.listener
	p.mu.Unlock()
	if listener == nil {
		return nil, NewTTransportException(NOT_OPEN, "No underlying server socket")
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, NewTTransportExceptionFromError(err)
	}

	cfg := newTConfiguration()
	cfg.SocketTimeout = p.clientTimeout
	return NewTSocketFromConnConf(uuid.NewUUID().Int64(), conn, cfg), nil
}

func (p *TServerSocket) Serve(tc *TContext, conf *TConfiguration) (err error) {
	if err = p.Listen(); err == nil {
		for {
			if socket, err := p.Accept(); err == nil {
				go func() {
					if conf != nil {
						socket.SetTConfiguration(conf)
					}
					socket.On(tc)
				}()
			}
		}
	}
	return
}

func (p *TServerSocket) AcceptLoop(tc *TContext, conf *TConfiguration) (err error) {
	if p.listener == nil {
		return errors.New("the service is not listening")
	}
	for {
		if socket, err := p.Accept(); err == nil {
			go func() {
				if conf != nil {
					socket.SetTConfiguration(conf)
				}
				socket.On(tc)
			}()
		}
	}
	return
}

// IsListening Checks whether the socket is listening.
func (p *TServerSocket) IsListening() bool {
	return p.listener != nil
}

// Open Connects the socket, creating a new socket object if necessary.
func (p *TServerSocket) Open() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsListening() {
		return NewTTransportException(ALREADY_OPEN, "Server socket already open")
	}
	if l, err := net.Listen(p.addr.Network(), p.addr.String()); err != nil {
		return err
	} else {
		p.listener = l
	}
	return nil
}

func (p *TServerSocket) Addr() net.Addr {
	if p.listener != nil {
		return p.listener.Addr()
	}
	return p.addr
}

func (p *TServerSocket) Close() error {
	var err error
	p.mu.Lock()
	if p.IsListening() {
		err = p.listener.Close()
		p.listener = nil
	}
	p.mu.Unlock()
	return err
}

func (p *TServerSocket) Interrupt() error {
	p.mu.Lock()
	p.interrupted = true
	p.mu.Unlock()
	p.Close()
	return nil
}
