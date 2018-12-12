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

// Package tentacle implements the Tentacle manager.
//
// To use this package:
//
//   1. Set the AppName and AppVersion variables.
//   2. Call LoadConfigure() to get the Configure.
//   3. Call tentacle.Start() to start Tentacle. You get back
//      an Instance, on which you can call Restart() to
//      restart it or Stop() to stop it.
//
// You should call Wait() on your instance to wait for
// all servers to quit before your process exits.
package tentacle

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/jiusanzhou/tentacle/pkg/conn"
	"github.com/jiusanzhou/tentacle/pkg/protocol"
)

// Configurable application parameters
var (
	// AppName is the name of the application.
	AppName string

	// AppVersion is the version of the application.
	AppVersion string

	// Quiet mode will not show any informative output on initialization.
	Quiet bool

	// PidFile is the path to the pid file to create.
	PidFile string

	// GracefulTimeout is the maximum duration of a graceful shutdown.
	GracefulTimeout time.Duration

	// isUpgrade will be set to true if this process
	// was started as part of an upgrade, where a parent
	// Caddy process started this one.
	isUpgrade = os.Getenv("TENTACLE__UPGRADE") == "1"

	// started will be set to true when the first
	// instance is started; it never gets set to
	// false after that.
	started bool

	// mu protects the variables 'isUpgrade' and 'started'.
	mu sync.Mutex
)

// Instance is a connection of
type Instance struct {
	// context is the context created for this instance,
	// used to coordinate the setting up of the server type
	context Context

	manager *Manager

	protocol protocol.Protocol

	conns chan net.Conn
}

// Start function is try to start the tentacle instance
func (inst *Instance) Start() error {

	return nil
}

func NewInstance(conn conn.Conn) *Instance {
	inst := &Instance{}

	// fork handle from manager
	inst.protocol = inst.manager.protocol.New()

	// register cmd handler
	inst.protocol.OnCommand(NewCmdHandler(inst.context))

	// register packet handler
	// TODO:

	// register event handler
	// TODO:

	return inst
}

func init() {

	// setup all functions
	setup()
}
