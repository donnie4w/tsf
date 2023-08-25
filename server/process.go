/**
 * Copyright 2023 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */

package server

import (
	"sync/atomic"

	. "github.com/donnie4w/gofer/buffer"
	. "github.com/donnie4w/tsf/packet"
	"github.com/donnie4w/tsf/util"
)

func Process(socket *TSocket, processPacKet func(socket *TSocket, pkt *Packet) error) (err error) {
	defer recover()
	defer socket.Close()
	for socket.conn.isValid() {
		bufhead := NewBuffer()
		headBit := int64(4)
		if socket.cfg.Packet64Bits {
			headBit = 8
		}
		if err = readsocket(socket, headBit, bufhead); err != nil {
			break
		}
		var ln int64
		if socket.cfg.Packet64Bits {
			ln = util.BytesToInt64(bufhead.Bytes())
		} else {
			ln = int64(util.BytesToInt32(bufhead.Bytes()))
		}
		buf := NewBuffer()
		if err = readsocket(socket, ln, buf); err == nil {
			pkt := Wrap(buf)
			pkt.Len = int(ln)
			if socket.cfg.Sync {
				processPacKet(socket, pkt)
			} else {
				go func() {
					defer recover()
					processPacKet(socket, pkt)
				}()
			}
		} else {
			break
		}
	}
	return
}

func readsocket(socket *TSocket, ln int64, buf *Buffer) (err error) {
	bs := make([]byte, ln)
	return _readsocket(socket, ln, buf, bs)
}

func _readsocket(socket *TSocket, ln int64, buf *Buffer, bs []byte) (err error) {
	var i int
	if i, err = socket.Read(bs); err == nil {
		buf.Write(bs[:i])
		if i < int(ln) {
			_readsocket(socket, ln-int64(i), buf, bs)
		}
	}
	return
}

func writeMerge(socket *TSocket) (int, error) {
	defer socket.mux.Unlock()
	ln, isbit64 := 4, socket.cfg.Packet64Bits
	if isbit64 {
		ln = 8
	}
	buf := NewBuffer()
	for bs := range socket._dataChan {
		bys := make([]byte, ln+len(bs))
		if isbit64 {
			copy(bys[:ln], util.Int64ToBytes(int64(len(bs))))
		} else {
			copy(bys[:ln], util.Int32ToBytes(int32(len(bs))))
		}
		copy(bys[ln:], bs)
		buf.Write(bys)
		if atomic.AddInt64(&socket._incount, -1) <= 0 {
			break
		}
	}

	ds := buf.Bytes()
	if socket.cfg.SnappyMergeData {
		ds = util.SnappyEncode(ds)
	}
	bys := make([]byte, ln+len(ds))
	if isbit64 {
		copy(bys[:ln], util.Int64ToBytes(int64(len(ds))))
	} else {
		copy(bys[:ln], util.Int32ToBytes(int32(len(ds))))
	}
	copy(bys[ln:], ds)
	return socket.Write(bys)
}

func ProcessMerge(socket *TSocket, processPacKet func(socket *TSocket, pkt *Packet) error) (err error) {
	defer recover()
	defer socket.Close()
	for socket.conn.isValid() {
		bufhead := NewBuffer()
		headBit := int64(4)
		if socket.cfg.Packet64Bits {
			headBit = 8
		}
		if err = readsocket(socket, headBit, bufhead); err != nil {
			break
		}
		var ln int64
		if socket.cfg.Packet64Bits {
			ln = util.BytesToInt64(bufhead.Bytes())
		} else {
			ln = int64(util.BytesToInt32(bufhead.Bytes()))
		}
		buf := NewBuffer()
		if err = readsocket(socket, ln, buf); err == nil {
			ds := buf.Bytes()
			if socket.cfg.SnappyMergeData {
				ds = util.SnappyDecode(ds)
			}
			_processMerge(ds, socket, processPacKet)
		} else {
			break
		}
	}
	return
}

func _processMerge(bs []byte, socket *TSocket, processPacKet func(socket *TSocket, pkt *Packet) error) {
	if len(bs) > 0 {
		var ln int64
		bit := int64(4)
		if socket.cfg.Packet64Bits {
			bit = 8
			ln = util.BytesToInt64(bs[:bit])
		} else {
			ln = int64(util.BytesToInt32(bs[:bit]))
		}
		pkt := &Packet{Len: int(ln), Buf: NewBufferBySlice(bs[bit : bit+ln])}
		if socket.cfg.Sync {
			processPacKet(socket, pkt)
		} else {
			go func() {
				defer recover()
				processPacKet(socket, pkt)
			}()
		}
		if nbs := bs[bit+ln:]; len(nbs) > 0 {
			_processMerge(nbs, socket, processPacKet)
		}
	}
}
