/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package packet

import (
	. "github.com/donnie4w/gofer/buffer"
)

type Packet struct {
	Len int32
	buf *Buffer
}

func Wrap(buf *Buffer) (p *Packet) {
	p = new(Packet)
	p.buf = buf
	return
}

func (this *Packet) ToBytes() []byte {
	return this.buf.Bytes()
}

func (this *Packet) Free() {
	this.buf.Free()
}
