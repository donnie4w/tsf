/**
 * Copyright 2023 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */

package server

import "sync"

type TsfSocket interface {
	_Incount () int64
	_SubAndGet() int64
	_Mux() *sync.Mutex
	_DataChan() chan []byte
	writebytes(buf []byte) (int, error)
	IsValid() bool
	Cfg() *TConfiguration
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	WriteWithMerge(buf []byte) (i int, err error)
	Interrupt() error 
	Close() error
}
