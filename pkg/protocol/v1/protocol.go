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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/jiusanzhou/tentacle/pkg/conn"
	"github.com/jiusanzhou/tentacle/pkg/protocol"
)

const (
	Version protocol.ProtocolVersion = 1
)

type protocolV1 struct {

	cmdHandler    protocol.CmdHandler
	eventHandler  protocol.EventHandler
	packetHandler protocol.PacketHandler

	mu sync.Mutex
}

func (v1 *protocolV1) Version() protocol.ProtocolVersion {
	return Version
}

func (v1 *protocolV1) IOLoop(conn conn.Conn) error {

	// first read frame try to handshake
	// we can distinguish tunnel's type: control/data
	data, idn, err := v1.readFrame(conn)
	if err != nil {
		return err
	}

	// we are control tunnel we need to handle command
	if idn == ControlIdentifier {
		// control tunnel
		// because we use a alone connection for control tunnel
		// so, we will just read the command all the time

		// the first command in the frame data
		// first we need to parse the command out from data

		// and the first command must be Hello command
		data, _, err = v1.readFrame(bytes.NewReader(data))

		var cmd uint32
		var response []byte
		var processed bool

		for {
			data, cmd, err = v1.readFrame(conn)
			if err != nil {
				return err
			}

			response, processed, err = v1.handleCmd(cmd, data)

			if !processed {
				// TODO: return unprocessed message
			}

			if err != nil {
				return err
			}

			if response != nil {
				// send cmd response
				// TODO:
			}
		}

	} else {
		// data tunnel
		for {
			data, idn, err = v1.readFrame(conn)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (v1 *protocolV1) OnCommand(handler protocol.CmdHandler) {
	var oldHandler = v1.cmdHandler
	v1.cmdHandler = protocol.CmdHandler(func(cmdType uint32, data []byte) ([]byte, bool, error) {
		if oldHandler != nil {
			response, processed, err := oldHandler(cmdType, data)
			if processed {
				return response, processed, err
			}
		}
		return handler(cmdType, data)
	})
}

func (v1 *protocolV1) OnEvent(handler func() error) {

}

func (v1 *protocolV1) OnPacket(handler func() error) {

}

func (v1 *protocolV1) New() protocol.Protocol {
	return &protocolV1{}
}

// TODO: set read timeout
func (v1 *protocolV1) readFrame(conn io.Reader) (data []byte, idn uint32, err error) {

	var (
		// TODO: use bytes pool or remove out of loop
		lenBytes = make([]byte, 4)
		idnBytes = make([]byte, 4)

		len uint32
	)

	// read length first
	err = fill(conn, lenBytes)
	if err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = fmt.Errorf("failed to read length from tunnel - %s", err)
		}
		return
	}

	// parse length
	len = binary.BigEndian.Uint32(lenBytes)

	// read identifier
	err = fill(conn, idnBytes)
	if err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = fmt.Errorf("failed to read identifier from tunnel - %s", err)
		}
	}

	// parse identifier
	idn = binary.BigEndian.Uint32(idnBytes)

	// read all data out
	// TODO: use buffer pool and if len > MaxByteSize
	data = make([]byte, len)
	err = fill(conn, data)

	return

}

func (v1 *protocolV1) ctlLoop() {

}

func (v1 *protocolV1) dataLoop() {

}

func init() {
	protocol.RegisterProtocol(&protocolV1{})
}
