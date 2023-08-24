/**
 * Copyright 2023 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */

package server

import (
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
			go func() {
				defer recover()
				processPacKet(socket, pkt)
			}()
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
