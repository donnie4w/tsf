package packet

import (
	gobuffer "github.com/donnie4w/gofer/pool/buffer"
)

var BytesPool = gobuffer.NewPool[[]byte](func() *[]byte {
	b := make([]byte, 128)
	return (*[]byte)(&b)
}, func(b *[]byte) { *b = (*b)[:0] })
