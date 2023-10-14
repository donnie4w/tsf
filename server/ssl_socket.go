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

package server

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/util"
	. "github.com/donnie4w/tsf/packet"
)

type TSSLSocket struct {
	conn *socketConn
	// hostPort contains host:port (e.g. "asdf.com:12345"). The field is
	// only valid if addr is nil.
	hostPort string
	// addr is nil when hostPort is not "", and is only used when the
	// TSSLSocket is constructed from a net.Addr.
	addr net.Addr

	cfg *TConfiguration

	_dataChan chan []byte
	mux       *sync.Mutex
	_incount  int64
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

// Deprecated: Use NewTSSLSocketConf instead.
// func NewTSSLSocket(hostPort string, cfg *tls.Config) (*TSSLSocket, error) {
// 	return NewTSSLSocketConf(hostPort, &TConfiguration{
// 		TLSConfig: cfg,

// 		noPropagation: true,
// 	}), nil
// }

// Deprecated: Use NewTSSLSocketConf instead.
// func NewTSSLSocketTimeout(hostPort string, cfg *tls.Config, connectTimeout, socketTimeout time.Duration) (*TSSLSocket, error) {
// 	return NewTSSLSocketConf(hostPort, &TConfiguration{
// 		ConnectTimeout: connectTimeout,
// 		SocketTimeout:  socketTimeout,
// 		TLSConfig:      cfg,

// 		noPropagation: true,
// 	}), nil
// }

// NewTSSLSocketFromAddrConf creates a TSSLSocket from a net.Addr.
func NewTSSLSocketFromAddrConf(addr net.Addr, conf *TConfiguration) *TSSLSocket {
	return &TSSLSocket{
		addr: addr,
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

// Deprecated: Use NewTSSLSocketFromAddrConf instead.
// func NewTSSLSocketFromAddrTimeout(addr net.Addr, cfg *tls.Config, connectTimeout, socketTimeout time.Duration) *TSSLSocket {
// 	return NewTSSLSocketFromAddrConf(addr, &TConfiguration{
// 		ConnectTimeout: connectTimeout,
// 		SocketTimeout:  socketTimeout,
// 		TLSConfig:      cfg,

// 		noPropagation: true,
// 	})
// }

// NewTSSLSocketFromConnConf creates a TSSLSocket from an existing net.Conn.
func NewTSSLSocketFromConnConf(conn net.Conn, conf *TConfiguration) *TSSLSocket {
	return &TSSLSocket{
		conn: wrapSocketConn(conn),
		addr: conn.RemoteAddr(),
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

// Deprecated: Use NewTSSLSocketFromConnConf instead.
func NewTSSLSocketFromConnTimeout(conn net.Conn, cfg *tls.Config, socketTimeout time.Duration) *TSSLSocket {
	return NewTSSLSocketFromConnConf(conn, &TConfiguration{
		SocketTimeout: socketTimeout,
		TLSConfig:     cfg,

		// noPropagation: true,
	})
}
func (p *TSSLSocket) _Incount() int64        { return p._incount }
func (p *TSSLSocket) _SubAndGet() int64      { return atomic.AddInt64(&p._incount, -1) }
func (p *TSSLSocket) IsValid() bool          { return p.conn.isValid() }
func (p *TSSLSocket) Cfg() *TConfiguration   { return p.cfg }
func (p *TSSLSocket) _Mux() *sync.Mutex      { return p.mux }
func (p *TSSLSocket) _DataChan() chan []byte { return p._dataChan }

// SetTConfiguration implements TConfigurationSetter.
//
// It can be used to change connect and socket timeouts.
func (p *TSSLSocket) SetTConfiguration(conf *TConfiguration) {
	p.cfg = conf
}

// Sets the connect timeout
func (p *TSSLSocket) SetConnTimeout(timeout time.Duration) error {
	if p.cfg == nil {
		p.cfg = &TConfiguration{}
	}
	p.cfg.ConnectTimeout = timeout
	return nil
}

// Sets the socket timeout
func (p *TSSLSocket) SetSocketTimeout(timeout time.Duration) error {
	if p.cfg == nil {
		p.cfg = &TConfiguration{}
	}
	p.cfg.SocketTimeout = timeout
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

// Connects the socket, creating a new socket object if necessary.
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
	return nil
}

// Retrieve the underlying net.Conn
func (p *TSSLSocket) Conn() net.Conn {
	return p.conn
}

// Returns true if the connection is open
func (p *TSSLSocket) IsOpen() bool {
	return p.conn.IsOpen()
}

// Closes the socket.
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

func (p *TSSLSocket) Write(buf []byte) (int, error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	p.pushDeadline(false, true)
	return p.conn.Write(buf)
}

func (p *TSSLSocket) WriteWithLen(buf []byte) (int, error) {
	ln := 4
	if p.cfg.Packet64Bits {
		ln = 8
	}
	bys := make([]byte, ln+len(buf))
	if p.cfg.Packet64Bits {
		copy(bys[:ln], util.Int64ToBytes(int64(len(buf))))
	} else {
		copy(bys[:ln], util.Int32ToBytes(int32(len(buf))))
	}
	copy(bys[ln:], buf)
	return p.Write(bys)
}

func (p *TSSLSocket) WriteWithMerge(buf []byte) (i int, err error) {
	if p._dataChan == nil {
		p.mux.Lock()
		if p._dataChan == nil {
			p._dataChan = make(chan []byte, 1<<13)
		}
		p.mux.Unlock()
	}
	p._dataChan <- buf
	atomic.AddInt64(&p._incount, 1)
	if p.mux.TryLock() {
		i, err = writeMerge(p)
	}
	return
}

func (p *TSSLSocket) Flush(ctx context.Context) error {
	return nil
}

func (p *TSSLSocket) Interrupt() error {
	if !p.conn.isValid() {
		return nil
	}
	return p.conn.Close()
}

func (p *TSSLSocket) RemainingBytes() (num_bytes uint64) {
	const maxSize = ^uint64(0)
	return maxSize // the truth is, we just don't know unless framed is used
}

var _ TConfigurationSetter = (*TSSLSocket)(nil)

func (p *TSSLSocket) Process(fn func(pkt *Packet) error) {
	Process(p, func(socket TsfSocket, pkt *Packet) error {
		return fn(pkt)
	})
}

func (p *TSSLSocket) ProcessMerge(fn func(pkt *Packet) error) {
	ProcessMerge(p, func(socket TsfSocket, pkt *Packet) error {
		return fn(pkt)
	})
}
