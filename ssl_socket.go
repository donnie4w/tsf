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

package tsf

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/donnie4w/gofer/uuid"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/util"
)

type TSSLSocket struct {
	id int64

	conn *socketConn
	// hostPort contains host:port (e.g. "asdf.com:12345"). The field is
	// only valid if addr is nil.
	hostPort string
	// addr is nil when hostPort is not "", and is only used when the
	// TSSLSocket is constructed from a net.Addr.
	addr     net.Addr
	cfg      *TConfiguration
	mgChan   chan []byte
	mux      *sync.Mutex
	inputNum atomic.Int64
	mgLock   atomic.Int32
	ctx      context.Context
}

// NewTSSLSocketConf creates a net.Conn-backed TTransport, given a host and port.
//
// Example:
//
//	trans := thrift.NewTSSLSocketConf("localhost:9090", &TConfiguration{
//	    ConnectTimeout: time.Second, // Use 0 for no timeout
//	    SocketTimeout:  time.Second, // Use 0 for no timeout
//
//	    TLSConfig: &tls.Config{
//	        // Fill in tls config here.
//	    }
//	})
func NewTSSLSocketConf(hostPort string, conf *TConfiguration) *TSSLSocket {
	if cfg := conf.GetTLSConfig(); cfg != nil && cfg.MinVersion == 0 {
		cfg.MinVersion = tls.VersionTLS10
	}
	return &TSSLSocket{
		hostPort: hostPort,
		cfg:      conf,
		mux:      &sync.Mutex{},
	}
}

// NewTSSLSocketFromAddrConf creates a TSSLSocket from a net.Addr.
func NewTSSLSocketFromAddrConf(addr net.Addr, conf *TConfiguration) *TSSLSocket {
	return &TSSLSocket{
		addr: addr,
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

// NewTSSLSocketFromConnConf creates a TSSLSocket from an existing net.Conn.
func NewTSSLSocketFromConnConf(id int64, conn net.Conn, conf *TConfiguration) *TSSLSocket {
	return &TSSLSocket{
		id:   id,
		conn: wrapSocketConn(conn),
		addr: conn.RemoteAddr(),
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

func (p *TSSLSocket) ID() int64             { return p.id }
func (p *TSSLSocket) IsValid() bool         { return p.conn.isValid() }
func (p *TSSLSocket) Cfg() *TConfiguration  { return p.cfg }
func (p *TSSLSocket) dataChan() chan []byte { return p.mgChan }
func (p *TSSLSocket) pendNumber() int64     { return p.inputNum.Load() }
func (p *TSSLSocket) pendSub() int64        { return p.inputNum.Add(-1) }

// SetTConfiguration implements TConfigurationSetter.
//
// It can be used to change connect and socket timeouts.
func (p *TSSLSocket) SetTConfiguration(conf *TConfiguration) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.cfg = conf
}

// SetContext sets the context for the TSocket instance.
// This method is thread-safe in a multi-threaded environment because it uses a mutex to protect the write operation of the context.
//
// Parameters:
//
//	ctx context.Context: The new context to set.
func (p *TSSLSocket) SetContext(ctx context.Context) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.ctx = ctx
}

// GetContext returns the current context for the TSocket instance.
// This method is thread-safe because it only reads the context, which is protected by a mutex in the SetContext method.
//
// Returns:
//
//	context.Context: The current context.
func (p *TSSLSocket) GetContext() context.Context {
	return p.ctx
}

// SetConnTimeout the connect timeout
func (p *TSSLSocket) SetConnTimeout(timeout time.Duration) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.ConnectTimeout = timeout
	return nil
}

// SetSocketTimeout the socket timeout
func (p *TSSLSocket) SetSocketTimeout(timeout time.Duration) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.SocketTimeout = timeout
	return nil
}

// SetSnappy Whether compressed merge data
func (p *TSSLSocket) SetSnappy(SnappyCompress bool) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.Snappy = SnappyCompress
	return nil
}

// SetBits64 Whether the packet size is 64-bit binary
func (p *TSSLSocket) SetBits64(bit64 bool) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.Bit64 = bit64
	return nil
}

func (p *TSSLSocket) pushDeadline(read, write bool) {
	var t time.Time
	if timeout := p.cfg.GetSocketTimeout(); timeout > 0 {
		t = time.Now().Add(time.Duration(timeout))
	}
	if read && write {
		p.conn.SetDeadline(t)
	} else if read {
		p.conn.SetReadDeadline(t)
	} else if write {
		p.conn.SetWriteDeadline(t)
	}
}

// Open Connects the socket, creating a new socket object if necessary.
func (p *TSSLSocket) Open() error {
	var err error
	// If we have a hostname, we need to pass the hostname to tls.Dial for
	// certificate hostname checks.
	if p.hostPort != "" {
		if p.conn, err = createSocketConnFromReturn(tls.DialWithDialer(
			&net.Dialer{
				Timeout: p.cfg.GetConnectTimeout(),
			},
			"tcp",
			p.hostPort,
			p.cfg.GetTLSConfig(),
		)); err != nil {
			return &tTransportException{
				typeId: NOT_OPEN,
				err:    err,
				msg:    err.Error(),
			}
		}
	} else {
		if p.conn.isValid() {
			return NewTTransportException(ALREADY_OPEN, "Socket already connected.")
		}
		if p.addr == nil {
			return NewTTransportException(NOT_OPEN, "Cannot open nil address.")
		}
		if len(p.addr.Network()) == 0 {
			return NewTTransportException(NOT_OPEN, "Cannot open bad network name.")
		}
		if len(p.addr.String()) == 0 {
			return NewTTransportException(NOT_OPEN, "Cannot open bad address.")
		}
		if p.conn, err = createSocketConnFromReturn(tls.DialWithDialer(
			&net.Dialer{
				Timeout: p.cfg.GetConnectTimeout(),
			},
			p.addr.Network(),
			p.addr.String(),
			p.cfg.GetTLSConfig(),
		)); err != nil {
			return &tTransportException{
				typeId: NOT_OPEN,
				err:    err,
				msg:    err.Error(),
			}
		}
	}
	p.id = uuid.NewUUID().Int64()
	return nil
}

// Conn Retrieve the underlying net.Conn
func (p *TSSLSocket) Conn() net.Conn {
	return p.conn
}

// IsOpen Returns true if the connection is open
func (p *TSSLSocket) IsOpen() bool {
	return p.conn.IsOpen()
}

// Close Closes the socket.
func (p *TSSLSocket) Close() error {
	return p.conn.Close()
}

func (p *TSSLSocket) Read(buf []byte) (int, error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	p.pushDeadline(true, false)
	// NOTE: Calling any of p.IsOpen, p.conn.read0, or p.conn.IsOpen between
	// p.pushDeadline and p.conn.Read could cause the deadline set inside
	// p.pushDeadline being reset, thus need to be avoided.
	n, err := p.conn.Read(buf)
	return n, NewTTransportExceptionFromError(err)
}

func (p *TSSLSocket) writeHandle(buf []byte) (i int, err error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	if err = overMessageSize(buf, p.cfg); err != nil {
		return
	}
	p.pushDeadline(false, true)
	return p.conn.Write(buf)
}

func (p *TSSLSocket) Write(buf []byte) (i int, err error) {
	if err = overMessageSize(buf, p.cfg); err != nil {
		return
	}
	ln := 4
	if p.cfg.Bit64 {
		ln = 8
	}
	bys := make([]byte, ln+len(buf))
	if p.cfg.Bit64 {
		copy(bys[:ln], util.Int64ToBytes(int64(len(buf))))
	} else {
		copy(bys[:ln], util.Int32ToBytes(int32(len(buf))))
	}
	copy(bys[ln:], buf)
	return p.writeHandle(bys)
}

func (p *TSSLSocket) WriteWithMerge(buf []byte) (i int, err error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}

	if p.mgChan == nil {
		p.mux.Lock()
		if p.mgChan == nil {
			p.mgChan = make(chan []byte, p.cfg.GetBlockedLimit())
		}
		p.mux.Unlock()
	}

	if p.cfg.GetMergoMode() == REJECTION {
		if p.inputNum.Load() >= p.cfg.GetBlockedLimit() {
			err = errors.New("too much blocking data")
			return
		}
	}

	if err = overMessageSize(buf, p.cfg); err != nil {
		return
	}
	p.mgChan <- buf
	p.inputNum.Add(1)
	if p.mgLock.CompareAndSwap(0, 1) {
		defer p.mgLock.Store(0)
		i, err = writeMerge(p)
	}
	return
}

func (p *TSSLSocket) Interrupt() error {
	if !p.conn.isValid() {
		return nil
	}
	return p.conn.Close()
}

func (p *TSSLSocket) Process(fn func(pkt *Packet) error) error {
	return Process(p, func(socket TsfSocket, pkt *Packet) error {
		return fn(pkt)
	})
}

func (p *TSSLSocket) ProcessMerge(fn func(pkt *Packet) error) error {
	return ProcessMerge(p, func(socket TsfSocket, pkt *Packet) error {
		return fn(pkt)
	})
}

func (p *TSSLSocket) On(tc *TContext) (err error) {
	defer recoverable(&err)
	if tc.OnClose != nil {
		defer tc.OnClose(p)
	}
	if tc.OnOpen != nil {
		go func() {
			defer recoverable(&err)
			tc.OnOpen(p)
		}()
	}
	if tc.OnOpenSync != nil {
		tc.OnOpenSync(p)
	}
	if p.cfg != nil && p.cfg.ProcessMerge && tc.Handler != nil {
		err = p.ProcessMerge(func(pkt *Packet) (e error) {
			defer recoverable(&e)
			return tc.Handler(p, pkt)
		})
	} else if tc.Handler != nil {
		err = p.Process(func(pkt *Packet) (e error) {
			defer recoverable(&e)
			return tc.Handler(p, pkt)
		})
	}
	if err != nil && tc.OnError != nil {
		tc.OnError(err, p)
	}
	return
}
