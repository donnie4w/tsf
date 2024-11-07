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
	"crypto/tls"
	"time"
)

// Default TConfiguration values.
const (
	DEFAULT_MAX_MESSAGE_SIZE  = 100 * 1024 * 1024
	DEFAULT_CONNECT_TIMEOUT   = 0
	DEFAULT_SOCKET_TIMEOUT    = 0
	DEFAULT_MERGECHANNEL_SIZE = 1 << 17
)

type MERGEMODE byte

const (
	BLOCKED   MERGEMODE = 0
	REJECTION MERGEMODE = 1
)

// TConfiguration is a configuration structure for a TCP service, defining various parameters required for its operation.
type TConfiguration struct {
	// Bit64 specifies whether the TCP packet header should use a 64-bit binary representation. By default,
	// the TCP header uses a 32-bit representation. Setting this to true enables 64-bit headers, which may
	// impact performance and compatibility.
	Bit64 bool

	// Snappy indicates whether to use the Snappy compression algorithm for transmitting data. Snappy is a fast
	// compression algorithm that provides a good balance between compression ratio and speed, making it suitable
	// for scenarios where large amounts of data need to be compressed and transmitted efficiently.
	Snappy bool

	// SyncProcess determines whether data processing is synchronous or asynchronous. By default, data processing
	// is asynchronous. If set to true, all data processing operations will be executed synchronously in the main
	// thread, which may increase latency but simplifies program logic.
	SyncProcess bool

	// MaxMessageSize specifies the maximum allowed size (in bytes) for a single message. If this value is less
	// than or equal to 0, the system's default maximum message size will be used. This setting helps prevent
	// memory overflow issues caused by receiving excessively large messages.
	MaxMessageSize int32

	// ConnectTimeout sets the timeout duration for attempting to establish a connection. If a connection cannot
	// be established within this time, the operation will fail and return an error. A reasonable timeout setting
	// prevents long waits for invalid connections.
	ConnectTimeout time.Duration

	// SocketTimeout defines the duration that read or write operations can wait before timing out. Once the timeout
	// is reached, the current operation will be canceled and an error will be returned. This is crucial for ensuring
	// service responsiveness and avoiding deadlocks.
	SocketTimeout time.Duration

	// ProcessMerge indicates whether to enable data merging during processing. When multiple small data packets
	// arrive consecutively, enabling this option can merge them into a larger packet for more efficient processing,
	// reducing network overhead and improving performance.
	ProcessMerge bool

	// MergoMode determines the behavior when the merge channel reaches its maximum capacity. There are two modes:
	// blocking mode and rejection mode. In blocking mode, new data will be blocked until there is enough space.
	// In rejection mode, new data will be rejected and an error will be returned. The choice of mode depends on
	// the specific requirements of the application.
	MergoMode MERGEMODE

	// MergeChannelSize defines the maximum capacity of the merge channel. This value determines the amount of data
	// that can be temporarily stored during the data merging process. Adjusting this value appropriately can help
	// optimize service performance, especially in high-concurrency environments.
	MergeChannelSize int64

	// TLSConfig provides the TLS (Transport Layer Security) configuration for creating secure TCP connections.
	// By setting this field, the TCP service can support encrypted communication, enhancing data security.
	// If TLS is not required, this field can be left empty.
	TLSConfig *tls.Config
}

// newTConfiguration creates and returns a new TConfiguration instance with default values for MaxMessageSize and MergeChannelSize.
func newTConfiguration() *TConfiguration {
	return &TConfiguration{
		MaxMessageSize:   DEFAULT_MAX_MESSAGE_SIZE,
		MergeChannelSize: DEFAULT_MERGECHANNEL_SIZE,
	}
}

// GetMaxMessageSize returns the maximum message size that the implementation should follow.
// If tc is nil or MaxMessageSize is less than or equal to 0, DEFAULT_MAX_MESSAGE_SIZE is returned.
// This method is nil-safe.
func (tc *TConfiguration) GetMaxMessageSize() int32 {
	if tc == nil || tc.MaxMessageSize <= 0 {
		return DEFAULT_MAX_MESSAGE_SIZE
	}
	return tc.MaxMessageSize
}

// GetConnectTimeout returns the connect timeout that should be used by TSocket and TSSLSocket.
// If tc is nil or ConnectTimeout is less than 0, DEFAULT_CONNECT_TIMEOUT is returned.
// This method is nil-safe.
func (tc *TConfiguration) GetConnectTimeout() time.Duration {
	if tc == nil || tc.ConnectTimeout < 0 {
		return DEFAULT_CONNECT_TIMEOUT
	}
	return tc.ConnectTimeout
}

// GetSocketTimeout returns the socket timeout that should be used by TSocket and TSSLSocket.
// If tc is nil or SocketTimeout is less than 0, DEFAULT_SOCKET_TIMEOUT is returned.
// This method is nil-safe.
func (tc *TConfiguration) GetSocketTimeout() time.Duration {
	if tc == nil || tc.SocketTimeout < 0 {
		return DEFAULT_SOCKET_TIMEOUT
	}
	return tc.SocketTimeout
}

// GetTLSConfig returns the TLS configuration that should be used by TSSLSocket.
// If tc is nil, nil is returned.
// This method is nil-safe.
func (tc *TConfiguration) GetTLSConfig() *tls.Config {
	if tc == nil {
		return nil
	}
	return tc.TLSConfig
}

// GetBlockedLimit returns the maximum channel length.
// If tc is nil or MergeChannelSize is less than or equal to 0, DEFAULT_MERGECHANNEL_SIZE is returned.
// This method is nil-safe.
func (tc *TConfiguration) GetBlockedLimit() int64 {
	if tc == nil || tc.MergeChannelSize <= 0 {
		return DEFAULT_MERGECHANNEL_SIZE
	}
	return tc.MergeChannelSize
}

// GetMergoMode returns the merge mode that should be used.
// If tc is nil, BLOCKED is returned.
// This method is nil-safe.
func (tc *TConfiguration) GetMergoMode() MERGEMODE {
	if tc == nil {
		return BLOCKED
	}
	return tc.MergoMode
}
