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
	"errors"
	"time"
)

// defaultTimeout is the default timeout duration used if no timeout is specified.
const defaultTimeout = 30 * time.Second

// TsfConfig holds the configuration settings for the Tsf server.
type TsfConfig struct {
	// ListenAddr is the address on which the server will listen for incoming connections.
	ListenAddr string

	// Timeout is the timeout duration for the server operations.
	Timeout time.Duration

	// TConfiguration is the configuration for the TCP socket.
	TConfiguration *TConfiguration
}

// Tsf represents the main structure for the Tsf server.
type Tsf struct {
	// tss is the underlying TsfSocketServer instance.
	tss TsfSocketServer

	// tsfConfig holds the configuration for the Tsf server.
	tsfConfig *TsfConfig

	// close indicates whether the server has been closed.
	close bool

	// tContext defines the context for handling TCP socket events and operations.
	tContext *TContext
}

// NewTsf creates and returns a new Tsf instance with the provided configuration.
// It initializes the server with the specified listen address, timeout, and configuration.
// If the configuration is invalid, it returns an error.
func NewTsf(tsfconfig *TsfConfig, tContext *TContext) (r *Tsf, err error) {
	if tsfconfig == nil {
		err = errors.New("tsf config is nil")
		return
	}

	if tsfconfig.ListenAddr == "" {
		err = errors.New("tsf listenAddr is empty")
		return
	}

	if tContext == nil {
		err = errors.New("tsf context is nil")
		return
	}

	if tsfconfig.Timeout <= 0 {
		tsfconfig.Timeout = defaultTimeout
	}

	if tsfconfig.TConfiguration == nil {
		tsfconfig.TConfiguration = &TConfiguration{}
	}

	var tss TsfSocketServer
	if tsfconfig.TConfiguration.TLSConfig != nil {
		tss, err = NewTSSLServerSocketTimeout(tsfconfig.ListenAddr, tsfconfig.TConfiguration.TLSConfig, tsfconfig.Timeout)
	} else {
		tss, err = NewTServerSocketTimeout(tsfconfig.ListenAddr, tsfconfig.Timeout)
	}
	if err != nil {
		return nil, err
	}
	r = &Tsf{tss: tss, tsfConfig: tsfconfig, tContext: tContext}
	return
}

// Serve starts the Tsf server and begins serving incoming connections.
// It uses the provided TContext and TConfiguration to handle the connections.
// If the server is already closed, it returns an error.
func (t *Tsf) Serve() error {
	if t.tsfConfig == nil {
		return errors.New("tsf config is nil")
	}
	err := t.tss.Serve(t.tContext, t.tsfConfig.TConfiguration)
	if t.close {
		return errors.New("tsf server closed")
	} else {
		return err
	}
}

// Listen starts listening for incoming connections on the specified address.
// It returns an error if the listen operation fails.
func (t *Tsf) Listen() error {
	if t.tsfConfig == nil {
		return errors.New("tsf config is nil")
	}
	return t.tss.Listen()
}

// AcceptLoop handles the loop for accepting incoming connections.
// It returns an error if the acceptance process fails.
func (t *Tsf) AcceptLoop() error {
	if t.tsfConfig == nil {
		return errors.New("tsf config is nil")
	}
	return t.tss.AcceptLoop(t.tContext, t.tsfConfig.TConfiguration)
}

// Close stops the Tsf server and closes all active connections.
// It sets the close flag to true and calls the Close method on the underlying TsfSocketServer.
func (t *Tsf) Close() error {
	t.close = true
	return t.tss.Close()
}
