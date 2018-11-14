/*
 * Copyright (c) 2018 wellwell.work, LLC by Zoe
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tentacle

import "net"

// TCPServer is a type that can listen and serve connections.
// A TCPServer must associate with exactly zero or one net.Listeners.
type TCPServer interface {
	// Listen starts listening by creating a new listener
	// and returning it. It does not start accepting
	// connections. For UDP-only servers, this method
	// can be a no-op that returns (nil, nil).
	Listen() (net.Listener, error)

	// Serve starts serving using the provided listener.
	// Serve must start the server loop nearly immediately,
	// or at least not return any errors before the server
	// loop begins. Serve blocks indefinitely, or in other
	// words, until the server is stopped. For UDP-only
	// servers, this method can be a no-op that returns nil.
	Serve(net.Listener) error
}

// UDPServer is a type that can listen and serve packets.
// A UDPServer must associate with exactly zero or one net.PacketConns.
type UDPServer interface {
	// ListenPacket starts listening by creating a new packetconn
	// and returning it. It does not start accepting connections.
	// TCP-only servers may leave this method blank and return
	// (nil, nil).
	ListenPacket() (net.PacketConn, error)

	// ServePacket starts serving using the provided packetconn.
	// ServePacket must start the server loop nearly immediately,
	// or at least not return any errors before the server
	// loop begins. ServePacket blocks indefinitely, or in other
	// words, until the server is stopped. For TCP-only servers,
	// this method can be a no-op that returns nil.
	ServePacket(net.PacketConn) error
}

// Server is a type that can listen and serve. It supports both
// TCP and UDP, although the UDPServer interface can be used
// for more than just UDP.
//
// If the server uses TCP, it should implement TCPServer completely.
// If it uses UDP or some other protocol, it should implement
// UDPServer completely. If it uses both, both interfaces should be
// fully implemented. Any unimplemented methods should be made as
// no-ops that simply return nil values.
type Server interface {
	TCPServer
	UDPServer
}