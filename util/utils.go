/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package util

func Int32ToBytes(n int32) (bs []byte) {
	bs = make([]byte, 4)
	for i := 0; i < 4; i++ {
		bs[i] = byte(n >> (8 * (3 - i)))
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
