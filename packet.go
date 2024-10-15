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
	"github.com/donnie4w/gofer/buffer"
)

type Packet struct {
	Len int
	Buf *buffer.Buffer
}

func Wrap(buf *buffer.Buffer) (p *Packet) {
	p = new(Packet)
	p.Buf = buf
	return
}

func (this *Packet) ToBytes() []byte {
	return this.Buf.Bytes()
}

func (this *Packet) Free() {
	this.Buf.Free()
}
