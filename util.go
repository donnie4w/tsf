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
	"fmt"
	"strconv"
)

func overMessageSize(buf []byte, cfg *TConfiguration) (err error) {
	mms := int(cfg.GetMaxMessageSize())
	if len(buf) > mms {
		return errors.New("message size too big: " + strconv.Itoa(len(buf)) + ", and the maximum allowed message size is " + strconv.Itoa(mms))
	}
	return
}

func recoverable(err *error) {
	if r := recover(); r != nil {
		if err != nil {
			*err = errors.New(fmt.Sprint(r))
		}
	}
}
