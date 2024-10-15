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
	"sync"
)

type tsfsocket interface {
	_Incount() int64
	_SubAndGet() int64
	_Mux() *sync.Mutex
	_DataChan() chan []byte
	writebytes(buf []byte) (int, error)
	ID() int64
	IsValid() bool
	Cfg() *TConfiguration
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	WriteWithMerge(buf []byte) (i int, err error)
	Interrupt() error
	Close() error
	SetTConfiguration(conf *TConfiguration)
}

type TsfSocket interface {
	ID() int64
	IsValid() bool
	Write([]byte) (int, error)
	WriteWithMerge(buf []byte) (i int, err error)
	Interrupt() error
	Close() error
	SetTConfiguration(conf *TConfiguration)
}

type TsfSocketServer interface {
	Listen() error
	Open() error
	Interrupt() error
	Close() error
	Serve(*TContext, *TConfiguration) error
}

type TContext struct {
	OnClose func(TsfSocket) error
	OnError func(error, TsfSocket) error
	OnOpen  func(TsfSocket) error
	Handler func(TsfSocket, *Packet) error
}
