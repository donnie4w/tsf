package server

import (
	"testing"
	"time"

	. "github.com/donnie4w/gofer/buffer"
	"github.com/donnie4w/simplelog/logging"
	"github.com/donnie4w/tsf/packet"
)

func TestServer(t *testing.T) {
	if serversocket, err := NewTServerSocketTimeout(":20001", 10*time.Second); err == nil {
		if err = serversocket.Listen(); err == nil {
			for {
				if socket, err := serversocket.Accept(); err == nil {
					go Process(socket, process)
				}
			}
		}
	}
}

func process(sock *TSocket, pkt *packet.Packet) (err error) {
	logging.Debug(string(pkt.ToBytes()))
	time.Sleep(3 * time.Second)
	sock.WriteWithLen(pkt.ToBytes())
	return
}

func TestSocket(t *testing.T) {
	sock := NewTSocketConf(":20001", &TConfiguration{ConnectTimeout: 10 * time.Second, SocketTimeout: 10 * time.Second})
	if err := sock.Open(); err == nil {
		bs := []byte("haha111")
		buf := BufPool.Get()
		buf.WriteInt32(len(bs))
		buf.Write(bs)
		sock.Write(buf.Bytes())
	}
	Process(sock, process)
}
