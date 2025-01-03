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
	"github.com/donnie4w/gofer/buffer"
	"github.com/donnie4w/gofer/util"
)

func Process(socket tsfsocket, processHandler func(socket TsfSocket, pkt *Packet) error) (err error) {
	defer recoverable(&err)
	defer socket.Close()
	for socket.IsValid() {
		headBit := 4
		var headBs []byte
		if socket.Cfg().Bit64 {
			headBit = 8
			var _headBs [8]byte
			headBs = _headBs[:]
		} else {
			var _headBs [4]byte
			headBs = _headBs[:]
		}
		if err = readStream(socket, headBs, headBit); err != nil {
			break
		}
		var ln int64
		if socket.Cfg().Bit64 {
			ln = util.BytesToInt64(headBs)
		} else {
			ln = int64(util.BytesToInt32(headBs))
		}
		if ln <= 0 {
			break
		}

		bodyBs := make([]byte, int(ln))
		if err = readStream(socket, bodyBs, int(ln)); err == nil {
			pkt := Wrap((*buffer.Buffer)(&bodyBs))
			pkt.Len = int(ln)
			if processHandler != nil {
				if socket.Cfg().SyncProcess {
					processHandler(socket, pkt)
				} else {
					go func() {
						defer recoverable(nil)
						processHandler(socket, pkt)
					}()
				}
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

func readStream(socket tsfsocket, bs []byte, ln int) (err error) {
	var i int
	if i, err = socket.Read(bs); err == nil {
		if i < ln {
			readStream(socket, bs[i:], ln-i)
		}
	}
	return
}

func writeMerge(socket tsfsocket) (int, error) {
START:
	isbit64 := socket.Cfg().Bit64
	buf := buffer.NewBuffer()
	if !socket.Cfg().Snappy {
		if isbit64 {
			buf.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
		} else {
			buf.Write([]byte{0, 0, 0, 0})
		}
	}
	for bs := range socket.dataChan() {
		if isbit64 {
			buf.Write(util.Int64ToBytes(int64(len(bs))))
		} else {
			buf.Write(util.Int32ToBytes(int32(len(bs))))
		}
		buf.Write(bs)
		if socket.pendSub() <= 0 {
			break
		}
	}
	var ds []byte
	if socket.Cfg().Snappy {
		se := util.SnappyEncode(buf.Bytes())
		if isbit64 {
			ds = append(util.Int64ToBytes(int64(len(se))), se...)
		} else {
			ds = append(util.Int32ToBytes(int32(len(se))), se...)
		}
	} else {
		ds = buf.Bytes()
		if isbit64 {
			copy(ds[:8], util.Int64ToBytes(int64(len(ds)-8)))
		} else {
			copy(ds[:4], util.Int32ToBytes(int32(len(ds)-4)))
		}
	}
	i, err := socket.writeHandle(ds)
	if err == nil && socket.pendNumber() > 0 {
		goto START
	}
	return i, err
}

func ProcessMerge(socket tsfsocket, processHandler func(socket TsfSocket, pkt *Packet) error) (err error) {
	defer recoverable(&err)
	defer socket.Close()
	for socket.IsValid() {
		var headBs []byte
		headBit := 4
		if socket.Cfg().Bit64 {
			headBit = 8
			var _headBs [8]byte
			headBs = _headBs[:]
		} else {
			var _headBs [4]byte
			headBs = _headBs[:]
		}
		//headBs := make([]byte, headBit)
		if err = readStream(socket, headBs, headBit); err != nil {
			break
		}
		var ln int64
		if socket.Cfg().Bit64 {
			ln = util.BytesToInt64(headBs)
		} else {
			ln = int64(util.BytesToInt32(headBs))
		}
		if ln <= 0 {
			break
		}
		bodyBs := make([]byte, ln)
		if err = readStream(socket, bodyBs, int(ln)); err == nil {
			if socket.Cfg().Snappy {
				bodyBs = util.SnappyDecode(bodyBs)
			}
			if bodyBs == nil || len(bodyBs) == 0 {
				break
			}
			_processMerge(bodyBs, socket, processHandler)
		} else {
			break
		}
	}
	return errors.New("socket close")
}

func _processMerge(bs []byte, socket tsfsocket, processHandler func(socket TsfSocket, pkt *Packet) error) {
START:
	if len(bs) > 0 {
		var ln int64
		bit := int64(4)
		if socket.Cfg().Bit64 {
			bit = 8
			ln = util.BytesToInt64(bs[:bit])
		} else {
			ln = int64(util.BytesToInt32(bs[:bit]))
		}
		pkt := &Packet{Len: int(ln), Buf: buffer.NewBufferBySlice(bs[bit : bit+ln])}
		if processHandler != nil {
			if socket.Cfg().SyncProcess {
				processHandler(socket, pkt)
			} else {
				go func() {
					defer recoverable(nil)
					processHandler(socket, pkt)
				}()
			}
		}

		if nbs := bs[bit+ln:]; len(nbs) > 0 {
			//_processMerge(nbs, socket, processHandler)
			bs = nbs
			goto START
		}
	}
}
