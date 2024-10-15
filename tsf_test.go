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
	"fmt"
	"testing"
)

func Test_tsf(t *testing.T) {
	cfg := &TsfConfig{ListenAddr: ":20001"}
	cfg.OnOpen = func(socket TsfSocket) error {
		fmt.Println("OnOpen:", socket.ID())
		return nil
	}
	cfg.Handler = func(socket TsfSocket, bytes []byte) error {
		fmt.Println(string(bytes))
		return nil
	}
	cfg.OnClose = func(socket TsfSocket) error {
		fmt.Println("OnClose:", socket.ID())
		return nil
	}
	cfg.OnError = func(err error, socket TsfSocket) error {
		fmt.Println("OnError:", socket.ID(), ",err:", err)
		return nil
	}
	s, err := NewTsf(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s.Serve()
}
