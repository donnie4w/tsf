/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
/**
 * Copyright 2023 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package server

import (
	"errors"
	"fmt"

	"github.com/donnie4w/simplelog/logging"
)

var log = logging.NewLogger().SetFormat(logging.FORMAT_DATE | logging.FORMAT_TIME)

func SetLog(on bool) {
	if on {
		log.SetLevel(logging.LEVEL_ERROR)
	} else {
		log.SetLevel(logging.LEVEL_OFF)
	}
}

func overMessageSize(buf []byte, cfg *TConfiguration) (err error) {
	mms := int(cfg.MaxMessageSize)
	if cfg.MaxMessageSize <= 0 {
		mms = DEFAULT_MAX_MESSAGE_SIZE
	}
	if len(buf) > mms {
		s := fmt.Sprint("tsf error, maxMessageSize:", mms, ",got ", len(buf))
		log.Error(s)
		return errors.New(s)
	}
	return
}
