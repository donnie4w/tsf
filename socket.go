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
	"errors"
	"github.com/donnie4w/gofer/uuid"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/util"
)

type TSocket struct {
	id       int64
	conn     *socketConn
	addr     net.Addr
	cfg      *TConfiguration
	mux      *sync.Mutex
	mgChan   chan []byte
	inputNum atomic.Int64
	mgLock   atomic.Int32
	ctx      context.Context
}

// tcpAddr is a naive implementation of net.Addr that does nothing extra.
type tcpAddr string

var _ net.Addr = tcpAddr("")

func (ta tcpAddr) Network() string {
	return "tcp"
}

func (ta tcpAddr) String() string {
	return string(ta)
}

// NewTSocketConf creates a net.Conn-backed TTransport, given a host and port.
//
// Example:
//
//	trans, err := thrift.NewTSocketConf("localhost:9090", &TConfiguration{
//	    ConnectTimeout: time.Second, // Use 0 for no timeout
//	    SocketTimeout:  time.Second, // Use 0 for no timeout
//	})
func NewTSocketConf(hostPort string, conf *TConfiguration) *TSocket {
	return NewTSocketFromAddrConf(tcpAddr(hostPort), conf)
}

// NewTSocketFromAddrConf creates a TSocket from a net.Addr
func NewTSocketFromAddrConf(addr net.Addr, conf *TConfiguration) *TSocket {
	return &TSocket{
		addr: addr,
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

// NewTSocketFromConnConf creates a TSocket from an existing net.Conn.
func NewTSocketFromConnConf(id int64, conn net.Conn, conf *TConfiguration) *TSocket {
	return &TSocket{
		id:   id,
		conn: wrapSocketConn(conn),
		addr: conn.RemoteAddr(),
		cfg:  conf,
		mux:  &sync.Mutex{},
	}
}

func (p *TSocket) ID() int64             { return p.id }
func (p *TSocket) IsValid() bool         { return p.conn.isValid() }
func (p *TSocket) Cfg() *TConfiguration  { return p.cfg }
func (p *TSocket) pendNumber() int64     { return p.inputNum.Load() }
func (p *TSocket) pendSub() int64        { return p.inputNum.Add(-1) }
func (p *TSocket) dataChan() chan []byte { return p.mgChan }

// SetTConfiguration implements TConfigurationSetter.
//
// It can be used to set connect and socket timeouts.
func (p *TSocket) SetTConfiguration(conf *TConfiguration) {
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
func (p *TSocket) SetContext(ctx context.Context) {
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
func (p *TSocket) GetContext() context.Context {
	return p.ctx
}

// SetConnTimeout Sets the connect timeout
func (p *TSocket) SetConnTimeout(timeout time.Duration) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.ConnectTimeout = timeout
	return nil
}

// SetSocketTimeout Sets the socket timeout
func (p *TSocket) SetSocketTimeout(timeout time.Duration) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.SocketTimeout = timeout
	return nil
}

// SetSnappy Whether compressed merge data
func (p *TSocket) SetSnappy(SnappyCompress bool) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.Snappy = SnappyCompress
	return nil
}

// SetBits64 Whether the packet size is 64-bit binary
func (p *TSocket) SetBits64(bit64 bool) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.Bit64 = bit64
	return nil
}

// SetSyncProcess Whether process is synchronize
func (p *TSocket) SetSyncProcess(isSync bool) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.cfg == nil {
		p.cfg = newTConfiguration()
	}
	p.cfg.SyncProcess = isSync
	return nil
}

func (p *TSocket) pushDeadline(read, write bool) {
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
func (p *TSocket) Open() error {
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
	var err error
	if p.conn, err = createSocketConnFromReturn(net.DialTimeout(
		p.addr.Network(),
		p.addr.String(),
		p.cfg.GetConnectTimeout(),
	)); err != nil {
		return &tTransportException{
			typeId: NOT_OPEN,
			err:    err,
			msg:    err.Error(),
		}
	}
	p.id = uuid.NewUUID().Int64()
	p.addr = p.conn.RemoteAddr()
	return nil
}

// Conn Retrieve the underlying net.Conn
func (p *TSocket) Conn() net.Conn {
	return p.conn
}

// IsOpen Returns true if the connection is open
func (p *TSocket) IsOpen() bool {
	return p.conn.IsOpen()
}

// Close the socket.
func (p *TSocket) Close() error {
	return p.conn.Close()
}

// Addr  Returns the remote address of the socket.
func (p *TSocket) Addr() net.Addr {
	return p.addr
}

func (p *TSocket) Read(buf []byte) (int, error) {
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

func (p *TSocket) writeHandle(buf []byte) (i int, err error) {
	if !p.conn.isValid() {
		return 0, NewTTransportException(NOT_OPEN, "Connection not open")
	}
	if err = overMessageSize(buf, p.cfg); err != nil {
		return
	}
	p.pushDeadline(false, true)
	return p.conn.Write(buf)
}

func (p *TSocket) Write(buf []byte) (i int, err error) {
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

func (p *TSocket) WriteWithMerge(buf []byte) (i int, err error) {
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

func (p *TSocket) Interrupt() error {
	if !p.conn.isValid() {
		return nil
	}
	return p.conn.Close()
}

func (p *TSocket) Process(fn func(pkt *Packet) error) error {
	return Process(p, func(socket TsfSocket, pkt *Packet) error {
		return fn(pkt)
	})
}

func (p *TSocket) ProcessMerge(fn func(pkt *Packet) error) error {
	return ProcessMerge(p, func(socket TsfSocket, pkt *Packet) error {
		return fn(pkt)
	})
}

func (p *TSocket) On(tc *TContext) (err error) {
	defer Recoverable(&err)
	if tc.OnClose != nil {
		defer func() {
			p.Close()
			tc.OnClose(p)
		}()
	}
	if tc.OnOpen != nil {
		tc.OnOpen(p)
	}
	if p.cfg != nil && p.cfg.ProcessMerge && tc.Handler != nil {
		err = p.ProcessMerge(func(pkt *Packet) (e error) {
			defer Recoverable(&e)
			return tc.Handler(p, pkt)
		})
	} else if tc.Handler != nil {
		err = p.Process(func(pkt *Packet) (e error) {
			defer Recoverable(&e)
			return tc.Handler(p, pkt)
		})
	}
	if err != nil && tc.OnError != nil {
		tc.OnError(err, p)
	}
	return
}
