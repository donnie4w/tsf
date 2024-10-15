// Copyright 2017 tsf Author. All Rights Reserved.
// email: donnie4w@gmail.com

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

package tsf

import (
	"crypto/tls"
	"time"
)

// Default TConfiguration values.
const (
	DEFAULT_MAX_MESSAGE_SIZE = 100 * 1024 * 1024
	DEFAULT_MAX_FRAME_SIZE   = 16384000

	DEFAULT_TBINARY_STRICT_READ  = false
	DEFAULT_TBINARY_STRICT_WRITE = true

	DEFAULT_CONNECT_TIMEOUT = 0
	DEFAULT_SOCKET_TIMEOUT  = 0

	DEFAULT_SOCKET_CHAN_LEN = 1 << 15
)

type TConfiguration struct {
	//Whether the packet size is 64-bit binary
	Packet64Bits bool

	//Whether compressed merge data
	SnappyMergeData bool
	//Whether process is synchronize
	SyncProcess bool
	// If <= 0, DEFAULT_MAX_MESSAGE_SIZE will be used instead.
	MaxMessageSize int32

	// If <= 0, DEFAULT_MAX_FRAME_SIZE will be used instead.
	//
	// Also if MaxMessageSize < MaxFrameSize,
	// MaxMessageSize will be used instead.
	//MaxFrameSize int32

	// Connect and socket timeouts to be used by TSocket and TSSLSocket.
	//
	// 0 means no timeout.
	//
	// If <0, DEFAULT_CONNECT_TIMEOUT and DEFAULT_SOCKET_TIMEOUT will be
	// used.
	ConnectTimeout time.Duration
	SocketTimeout  time.Duration

	// TLS config to be used by TSSLSocket.
	TLSConfig *tls.Config

	// Used internally by deprecated constructors, to avoid overriding
	// underlying TTransport/TProtocol's cfg by accidental propagations.
	//
	// For external users this is always false.
	// noPropagation bool
}

func newTConfiguration() *TConfiguration {
	return &TConfiguration{MaxMessageSize: DEFAULT_MAX_MESSAGE_SIZE}
}

// GetMaxMessageSize returns the max message size an implementation should
// follow.
//
// It's nil-safe. DEFAULT_MAX_MESSAGE_SIZE will be returned if tc is nil.
func (tc *TConfiguration) GetMaxMessageSize() int32 {
	if tc == nil || tc.MaxMessageSize <= 0 {
		return DEFAULT_MAX_MESSAGE_SIZE
	}
	return tc.MaxMessageSize
}

// GetMaxFrameSize returns the max frame size an implementation should follow.
//
// It's nil-safe. DEFAULT_MAX_FRAME_SIZE will be returned if tc is nil.
//
// If the configured max message size is smaller than the configured max frame
// size, the smaller one will be returned instead.
//func (tc *TConfiguration) GetMaxFrameSize() int32 {
//	if tc == nil {
//		return DEFAULT_MAX_FRAME_SIZE
//	}
//	maxFrameSize := tc.MaxFrameSize
//	if maxFrameSize <= 0 {
//		maxFrameSize = DEFAULT_MAX_FRAME_SIZE
//	}
//	if maxMessageSize := tc.GetMaxMessageSize(); maxMessageSize < maxFrameSize {
//		return maxMessageSize
//	}
//	return maxFrameSize
//}

// GetConnectTimeout returns the connect timeout should be used by TSocket and
// TSSLSocket.
//
// It's nil-safe. If tc is nil, DEFAULT_CONNECT_TIMEOUT will be returned instead.
func (tc *TConfiguration) GetConnectTimeout() time.Duration {
	if tc == nil || tc.ConnectTimeout < 0 {
		return DEFAULT_CONNECT_TIMEOUT
	}
	return tc.ConnectTimeout
}

// GetSocketTimeout returns the socket timeout should be used by TSocket and
// TSSLSocket.
//
// It's nil-safe. If tc is nil, DEFAULT_SOCKET_TIMEOUT will be returned instead.
func (tc *TConfiguration) GetSocketTimeout() time.Duration {
	if tc == nil || tc.SocketTimeout < 0 {
		return DEFAULT_SOCKET_TIMEOUT
	}
	return tc.SocketTimeout
}

// GetTLSConfig returns the tls config should be used by TSSLSocket.
//
// It's nil-safe. If tc is nil, nil will be returned instead.
func (tc *TConfiguration) GetTLSConfig() *tls.Config {
	if tc == nil {
		return nil
	}
	return tc.TLSConfig
}

// TConfigurationSetter is an optional interface TProtocol, TTransport,
// TProtocolFactory, TTransportFactory, and other implementations can implement.
//
// It's intended to be called during intializations.
// The behavior of calling SetTConfiguration on a TTransport/TProtocol in the
// middle of a message is undefined:
// It may or may not change the behavior of the current processing message,
// and it may even cause the current message to fail.
//
// Note for implementations: SetTConfiguration might be called multiple times
// with the same value in quick successions due to the implementation of the
// propagation. Implementations should make SetTConfiguration as simple as
// possible (usually just overwrite the stored configuration and propagate it to
// the wrapped TTransports/TProtocols).
type TConfigurationSetter interface {
	SetTConfiguration(*TConfiguration)
}

// PropagateTConfiguration propagates cfg to impl if impl implements
// TConfigurationSetter and cfg is non-nil, otherwise it does nothing.
//
// NOTE: nil cfg is not propagated. If you want to propagate a TConfiguration
// with everything being default value, use &TConfiguration{} explicitly instead.
// func PropagateTConfiguration(impl interface{}, cfg *TConfiguration) {
// 	if cfg == nil || cfg.noPropagation {
// 		return
// 	}

// 	if setter, ok := impl.(TConfigurationSetter); ok {
// 		setter.SetTConfiguration(cfg)
// 	}
// }