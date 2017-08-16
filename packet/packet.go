package packet

import (
	"bytes"
	"fmt"

	. "github.com/donnie4w/tsf/utils"
)

type PacketType int16

const (
	Err PacketType = iota
	Auth
	Service
	Ping
	Register
	AckService
	AckRegister
	AckPing
)

type Packet struct {
	PType     PacketType
	Body      []byte
	SeqId     int64
	Serviceid []byte
}

func Wrap(bs []byte) (p *Packet) {
	p = new(Packet)
	p.PType = PacketType(BytesToShort(bs[0:2]))
	p.SeqId = ByteToInt64(bs[2:10])
	p.Serviceid = bs[10:26]
	p.Body = bs[26:]
	return
}

func NewPacket() *Packet {
	return &Packet{Body: []byte{0}, Serviceid: make([]byte, 16)}
}

func (this *Packet) GetPacket() (bs []byte) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(IntToBytes(2 + 8 + int32(len(this.Body)+len(this.Serviceid))))
	buffer.Write(ShortToBytes(int16(this.PType)))
	buffer.Write(Int64ToBytes(this.SeqId))
	buffer.Write(this.Serviceid)
	buffer.Write(this.Body)
	bs = buffer.Bytes()
	return
}

func (this *Packet) SetBody(b []byte) {
	this.Body = b
}

func (this *Packet) ToString() string {
	return fmt.Sprint("PacketType:", this.PType, " seqId:", this.SeqId, " length serviceid:", len(this.Serviceid), " length body:", len(this.Body))
}
