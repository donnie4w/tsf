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
	"context"
	"net"
)

type tsfsocket interface {
	pendNumber() int64
	pendSub() int64
	dataChan() chan []byte
	writeHandle(buf []byte) (int, error)
	ID() int64
	IsValid() bool
	Cfg() *TConfiguration
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	WriteWithMerge(buf []byte) (i int, err error)
	Interrupt() error
	Close() error
	SetTConfiguration(conf *TConfiguration)
	Process(fn func(pkt *Packet) error) error
	ProcessMerge(fn func(pkt *Packet) error) error
	IsOpen() bool
	Open() error
	Conn() net.Conn
}

// TsfSocket defines the interface for a TCP socket with additional features and configurations.
type TsfSocket interface {
	// ID returns the unique identifier of the socket.
	ID() int64

	// IsValid returns true if the socket is valid and can be used for communication.
	IsValid() bool

	// Write writes the given byte slice to the socket.
	// It returns the number of bytes written and any error encountered.
	Write([]byte) (int, error)

	// WriteWithMerge writes the given byte slice to the socket with merging.
	// It returns the number of bytes written and any error encountered.
	WriteWithMerge(buf []byte) (int, error)

	// Interrupt interrupts the current operation on the socket.
	// It returns an error if the interruption fails.
	Interrupt() error

	// Close closes the socket.
	// It returns an error if the close operation fails.
	Close() error

	// SetTConfiguration sets the configuration for the socket.
	// It takes a pointer to a TConfiguration struct.
	SetTConfiguration(conf *TConfiguration)

	// Open opens the socket.
	// It returns an error if the open operation fails.
	Open() error

	// Process processes incoming packets using the provided function.
	// It returns an error if the processing fails.
	Process(fn func(pkt *Packet) error) error

	// ProcessMerge processes incoming packets using the provided function with merging.
	// It returns an error if the processing fails.
	ProcessMerge(fn func(pkt *Packet) error) error

	// Conn returns the underlying net.Conn object.
	Conn() net.Conn

	// IsOpen returns true if the socket is currently open.
	IsOpen() bool

	// Read reads data from the socket into the given byte slice.
	// It returns the number of bytes read and any error encountered.
	Read(buf []byte) (int, error)

	// On sets the event handlers for the TSocket using the provided TContext.
	// It allows setting callbacks for socket events such as opening, closing, and error handling,
	// as well as a handler for processing incoming packets.
	// If any of the callbacks are set, they will be called when the corresponding events occur.
	// This method returns an error if the operation fails.
	On(tc *TContext) (err error)

	// SetContext sets the context for the TSocket instance.
	// This method is thread-safe in a multi-threaded environment because it uses a mutex to protect the write operation of the context.
	//
	// Parameters:
	//
	//	ctx context.Context: The new context to set.
	SetContext(ctx context.Context)

	// GetContext returns the current context for the TSocket instance.
	// This method is thread-safe because it only reads the context, which is protected by a mutex in the SetContext method.
	//
	// Returns:
	//
	//	context.Context: The current context.
	GetContext() context.Context
}

// TsfSocketServer defines the interface for a TCP socket server with additional features and configurations.
type TsfSocketServer interface {
	// Listen starts listening for incoming connections on the specified address.
	// It returns an error if the listen operation fails.
	Listen() error

	// Open opens the server, preparing it to accept connections.
	// It returns an error if the open operation fails.
	Open() error

	// Interrupt interrupts the server, stopping any ongoing operations.
	// It returns an error if the interruption fails.
	Interrupt() error

	// Close closes the server, shutting down all active connections.
	// It returns an error if the close operation fails.
	Close() error

	// Serve serves incoming connections using the provided context and configuration.
	// It returns an error if the serving process fails.
	Serve(ctx *TContext, conf *TConfiguration) error
}

// TContext defines the context for handling TCP socket events and operations.
type TContext struct {
	// OnClose is a callback function that is called when a socket is closed.
	// It takes a TsfSocket as an argument and returns an error if the operation fails.
	OnClose func(TsfSocket) error

	// OnError is a callback function that is called when an error occurs on a socket.
	// It takes an error and a TsfSocket as arguments and returns an error if the operation fails.
	OnError func(error, TsfSocket) error

	// OnOpen is a callback function that is called when a socket is opened.
	// It takes a TsfSocket as an argument and returns an error if the operation fails.
	OnOpen func(TsfSocket) error

	// Handler is a callback function that is called to handle incoming packets.
	// It takes a TsfSocket and a Packet as arguments and returns an error if the operation fails.
	Handler func(TsfSocket, *Packet) error
}

// NewTsfSocketConf creates and returns a new TsfSocket instance based on the provided hostPort and configuration.
// If the configuration includes a non-nil TLSConfig, it returns a TSSLSocket instance for secure connections.
// Otherwise, it returns a standard TSocket instance.
// This function ensures that the appropriate socket type is created based on the configuration.
func NewTsfSocketConf(hostPort string, conf *TConfiguration) TsfSocket {
	if conf != nil && conf.GetTLSConfig() != nil {
		return NewTSSLSocketConf(hostPort, conf)
	}
	return NewTSocketConf(hostPort, conf)
}
