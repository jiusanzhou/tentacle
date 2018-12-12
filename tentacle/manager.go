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

import (
	"context"

	"github.com/jiusanzhou/tentacle/pkg/config"
	"github.com/jiusanzhou/tentacle/pkg/protocol"
)

type Manager struct {
	context  context.Context
	protocol protocol.Protocol
	config   *config.Configuration
}

func NewManager(conf *config.Configuration) *Manager {
	m := &Manager{
		context:  context.Background(),
		config:   conf,
		protocol: protocol.MustWithProtocol(conf.ProtocolVersion),
	}
	return m
}
