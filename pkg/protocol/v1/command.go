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

package v1

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	ControlIdentifier = 0x00
)

// control payload layout (as same as data tunnel)
// +---------------------------------------------------------------+
// |                        Length (32)                            |
// +---------------+---------------+-------------------------------+
// +-+-------------+---------------+-------------------------------+
// |                   Command Identifier (16)                     |
// +-+=============================================================+
// |                   Frame Payload (0...)                      ...
// +---------------------------------------------------------------+

const (
	// Hello cmd occurs while the client connect to server at first time.
	Hello uint32 = 0x01
	// HeartBeat cmd occurs in each heart beat interval.
	Heartbeat uint32 = 0x02
	// Metrics cmd occurs while server want to get metrics information from client.
	Metrics uint32 = 0x03
	// NewTunnel cmd is send from server to client to open a new data tunnel.
	Tunnel uint32 = 0x04
	// Transport is try to open a tunnel to pipe data from port to port.
	Transport uint32 = 0x05
	// Stats is for instance's states information.
	Stats uint32 = 0x06
	// CMD cmd is command line.
	CMD uint32 = 0x07
)

var (
	ErrProtocolVersionNotMatch = fmt.Errorf("server served protocol with version: %s", Version)
	ErrUnauthorized            = errors.New("unauthorized request, plz contact the admin")
)

func (v1 *protocolV1) handleCmd(cmdType uint32, data []byte) (response []byte, processed bool, err error) {

	// TODO: why we need to a lock for command?
	// Do we real need to handle only 1 command in the same time?
	v1.mu.Lock()
	defer v1.mu.Unlock()

	switch cmdType {
	case Hello:
		processed = true
		// the data must be the version of protocol and auth information
		// |-|-auth information-|
		// |1|                  |
		if binary.BigEndian.Uint16(data[:2]) != uint16(Version) {
			err = ErrProtocolVersionNotMatch
			return
		}

		// check the auth information
		if !v1.auth(data[2:]) {
			// TODO: failed callback
			err = ErrUnauthorized
			return
		}
		// TODO: success callback

		// TODO: send a token to client
		// client must use this token to access other commands or tunnel
		// client.WriteMsg()

	case Heartbeat:
		processed = true

	case Metrics:
		processed = true

	case Tunnel:
		processed = true

	case Transport:
		processed = true

	case Stats:
		processed = true

	case CMD:
		processed = true
	default:
		response, processed, err = v1.cmdHandler(cmdType, data)
	}

	return
}

func (v1 protocolV1) auth(data []byte) bool {
	// TODO: we need to get the instance and then call the auth method to get the result
	return true
}
