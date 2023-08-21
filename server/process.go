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

func Process(socket *TSocket, processPacKet func(socket *TSocket, pkt *Packet) error) {
	defer recover()
	for socket.conn.isValid() {
		bufhead := BufPool.Get()
		err := readsocket(socket, 4, bufhead)
		if err != nil {
			break
		}
		ln := util.BytesToInt32(bufhead.Bytes())
		bufhead.Free()
		buf := BufPool.Get()
		if err := readsocket(socket, ln, buf); err == nil {
			pkt := Wrap(buf)
			go func() {
				defer recover()
				processPacKet(socket, pkt)
			}()
		}
	}
}

func readsocket(socket *TSocket, ln int32, buf *Buffer) (err error) {
	bs := make([]byte, ln)
	var i int
	if i, err = socket.Read(bs); err == nil {
		buf.Write(bs)
		if i < int(ln) {
			readsocket(socket, ln-int32(i), buf)
		}
	}
	return
}
