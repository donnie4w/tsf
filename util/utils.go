/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package util

import "github.com/golang/snappy"

func Int32ToBytes(n int32) (bs []byte) {
	bs = make([]byte, 4)
	for i := 0; i < 4; i++ {
		bs[i] = byte(n >> (8 * (3 - i)))
	}
	return
}

func Int64ToBytes(n int64) (bs []byte) {
	bs = make([]byte, 8)
	for i := 0; i < 8; i++ {
		bs[i] = byte(n >> (8 * (7 - i)))
	}
	return
}

func BytesToInt32(bs []byte) (_r int32) {
	if len(bs) >= 4 {
		for i := 0; i < 4; i++ {
			_r = _r | int32(bs[i])<<(8*(3-i))
		}
	} else {
		bs4 := make([]byte, 4)
		for i, b := range bs {
			bs4[3-i] = b
		}
		_r = BytesToInt32(bs4)
	}
	return
}

func BytesToInt64(bs []byte) (_r int64) {
	if len(bs) >= 8 {
		for i := 0; i < 8; i++ {
			_r = _r | int64(bs[i])<<(8*(7-i))
		}
	} else {
		bs8 := make([]byte, 8)
		for i, b := range bs {
			bs8[7-i] = b
		}
		_r = BytesToInt64(bs8)
	}
	return
}

func SnappyEncode(bs []byte) (_r []byte) {
	return snappy.Encode(nil, bs)
}

func SnappyDecode(bs []byte) (_r []byte) {
	_r, _ = snappy.Decode(nil, bs)
	return
}
