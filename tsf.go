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

const (
	defaultTimeout = 30 * time.Second
)

type TsfConfig struct {
	ListenAddr     string
	Timeout        time.Duration
	TConfiguration *TConfiguration
	OnError        func(error, TsfSocket) error
	OnClose        func(TsfSocket) error
	OnOpen         func(TsfSocket) error
	Handler        func(TsfSocket, []byte) error
}

type Tsf struct {
	tserver   TsfSocketServer
	tsfConfig *TsfConfig
	close     bool
}

func NewTsf(tsfconfig *TsfConfig) (r *Tsf, err error) {
	if tsfconfig == nil {
		err = errors.New("tsf config is nil")
		return
	}

	if tsfconfig.ListenAddr == "" {
		err = errors.New("tsf listenAddr is empty")
		return
	}

	if tsfconfig.Timeout <= 0 {
		tsfconfig.Timeout = defaultTimeout
	}

	if tsfconfig.TConfiguration == nil {
		tsfconfig.TConfiguration = &TConfiguration{}
	}

	var tserver TsfSocketServer
	if tsfconfig.TConfiguration.TLSConfig != nil {
		tserver, err = NewTSSLServerSocketTimeout(tsfconfig.ListenAddr, tsfconfig.TConfiguration.TLSConfig, tsfconfig.Timeout)
	} else {
		tserver, err = NewTServerSocketTimeout(tsfconfig.ListenAddr, tsfconfig.Timeout)
	}
	if err != nil {
		return nil, err
	}
	r = &Tsf{tserver: tserver, tsfConfig: tsfconfig}
	return
}

func (t *Tsf) Serve() error {
	tc := &TContext{}
	tc.OnClose = t.tsfConfig.OnClose
	tc.OnOpen = t.tsfConfig.OnOpen
	tc.Handler = func(socket TsfSocket, p *Packet) error { return t.tsfConfig.Handler(socket, p.ToBytes()) }
	tc.OnError = t.tsfConfig.OnError
	err := t.tserver.Serve(tc, t.tsfConfig.TConfiguration)
	if t.close {
		return errors.New("tsf server closed")
	} else {
		return err
	}
}

func (t *Tsf) Stop() error {
	t.close = true
	return t.tserver.Close()
}
