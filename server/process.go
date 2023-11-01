/**
 * Copyright 2023 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */

package server

import (
	"errors"

	. "github.com/donnie4w/gofer/buffer"
	"github.com/donnie4w/gofer/util"
	. "github.com/donnie4w/tsf/packet"
)

func Process(socket TsfSocket, processPacKet func(socket TsfSocket, pkt *Packet) error) (err error) {
	defer recover()
	defer socket.Close()
	for socket.IsValid() {
		headBit := 4
		if socket.Cfg().Packet64Bits {
			headBit = 8
		}
		headBs := make([]byte, headBit)
		if err = readStream(socket, headBs, headBit); err != nil {
			break
		}
		var ln int64
		if socket.Cfg().Packet64Bits {
			ln = util.BytesToInt64(headBs)
		} else {
			ln = int64(util.BytesToInt32(headBs))
		}
		if ln <= 0 {
			break
		}

		bodyBs := make([]byte, int(ln))
		if err = readStream(socket, bodyBs, int(ln)); err == nil {
			pkt := Wrap((*Buffer)(&bodyBs))
			pkt.Len = int(ln)
			if socket.Cfg().SyncProcess {
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
	return errors.New("socket close")
}

// func readsocket(socket *TSocket, ln int64, buf *Buffer) (err error) {
// 	bs := make([]byte, ln)
// 	return _readsocket(socket, ln, buf, bs)
// }

// func _readsocket(socket *TSocket, ln int64, buf *Buffer, bs []byte) (err error) {
// 	var i int
// 	if i, err = socket.Read(bs); err == nil {
// 		buf.Write(bs[:i])
// 		if i < int(ln) {
// 			_readsocket(socket, ln-int64(i), buf, bs)
// 		}
// 	}
// 	return
// }

func readStream(socket TsfSocket, bs []byte, ln int) (err error) {
	var i int
	if i, err = socket.Read(bs); err == nil {
		if i < ln {
			readStream(socket, bs[i:], ln-i)
		}
	}
	return
}

func writeMerge(socket TsfSocket) (i int, err error) {
	defer socket._Mux().Unlock()
	return _writeMerge(socket)
}

func _writeMerge(socket TsfSocket) (i int, err error) {
	ln, isbit64 := 4, socket.Cfg().Packet64Bits
	if isbit64 {
		ln = 8
	}
	buf := NewBuffer()
	for bs := range socket._DataChan() {
		bys := make([]byte, ln+len(bs))
		if isbit64 {
			copy(bys[:ln], util.Int64ToBytes(int64(len(bs))))
		} else {
			copy(bys[:ln], util.Int32ToBytes(int32(len(bs))))
		}
		copy(bys[ln:], bs)
		buf.Write(bys)
		if socket._SubAndGet() <= 0 {
			break
		}
	}

	ds := buf.Bytes()
	if socket.Cfg().SnappyMergeData {
		ds = util.SnappyEncode(ds)
	}
	bys := make([]byte, ln+len(ds))
	if isbit64 {
		copy(bys[:ln], util.Int64ToBytes(int64(len(ds))))
	} else {
		copy(bys[:ln], util.Int32ToBytes(int32(len(ds))))
	}
	copy(bys[ln:], ds)
	i, err = socket.writebytes(bys)
	if socket._Incount() > 0 {
		return _writeMerge(socket)
	}
	return
}

func ProcessMerge(socket TsfSocket, processPacKet func(socket TsfSocket, pkt *Packet) error) (err error) {
	defer recover()
	defer socket.Close()
	for socket.IsValid() {
		headBit := 4
		if socket.Cfg().Packet64Bits {
			headBit = 8
		}
		headBs := make([]byte, headBit)
		if err = readStream(socket, headBs, headBit); err != nil {
			break
		}
		var ln int64
		if socket.Cfg().Packet64Bits {
			ln = util.BytesToInt64(headBs)
		} else {
			ln = int64(util.BytesToInt32(headBs))
		}
		if ln <= 0 {
			break
		}
		bodyBs := make([]byte, ln)
		if err = readStream(socket, bodyBs, int(ln)); err == nil {
			if socket.Cfg().SnappyMergeData {
				bodyBs = util.SnappyDecode(bodyBs)
			}
			if bodyBs == nil || len(bodyBs) == 0 {
				break
			}
			_processMerge(bodyBs, socket, processPacKet)
		} else {
			break
		}
	}
	return errors.New("socket close")
}

func _processMerge(bs []byte, socket TsfSocket, processPacKet func(socket TsfSocket, pkt *Packet) error) {
	if len(bs) > 0 {
		var ln int64
		bit := int64(4)
		if socket.Cfg().Packet64Bits {
			bit = 8
			ln = util.BytesToInt64(bs[:bit])
		} else {
			ln = int64(util.BytesToInt32(bs[:bit]))
		}
		pkt := &Packet{Len: int(ln), Buf: NewBufferBySlice(bs[bit : bit+ln])}
		if socket.Cfg().SyncProcess {
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
