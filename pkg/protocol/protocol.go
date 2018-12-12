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

package protocol

import (
	"fmt"

	"github.com/jiusanzhou/tentacle/pkg/conn"
)

var protocols = make(map[ProtocolVersion]Protocol)

type ProtocolVersion uint16

type (
	// CmdHandler is handler for control pipeline
	CmdHandler func(cmdType uint32, data []byte) ([]byte, bool, error)

	// PacketHandler is handler for data pipeline
	PacketHandler func(idn uint32, data []byte) error

	// EventHandler is handler for event while loop
	EventHandler func(eventType uint32, data []byte) error
)

type Protocol interface {
	Version() ProtocolVersion
	IOLoop(conn conn.Conn) error

	OnCommand(handler CmdHandler)
	OnEvent(func() error)
	OnPacket(func() error)

	New() Protocol
}

func RegisterProtocol(p Protocol) {
	protocols[p.Version()] = p
}

func MustWithProtocol(ver int) Protocol {
	ptc, ok := protocols[ProtocolVersion(ver)]
	if !ok {
		panic(fmt.Sprint("not such protocol version: %s", ver))
	}
	return ptc
}
