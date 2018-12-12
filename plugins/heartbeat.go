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

package plugins

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/jiusanzhou/tentacle/tentacle"
)

type Ping struct {
	time int64
}

func (ping *Ping) Unmarshal(data []byte) error {
	ping.time = int64(binary.BigEndian.Uint64(data))
	return nil
}

func (ping *Ping) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(ping.time))
	return b, nil
}

type Pong struct {
	time int64
}

func (pong *Pong) Unmarshal(data []byte) error {
	pong.time = int64(binary.BigEndian.Uint64(data))
	return nil
}

func (pong *Pong) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(pong.time))
	return b, nil
}

func heartbeat(ctx tentacle.Context, ping *Ping) (*Pong, error) {
	now := time.Now().UnixNano()
	fmt.Println("time delay:", now-ping.time)
	return &Pong{
		time: now,
	}, nil
}


