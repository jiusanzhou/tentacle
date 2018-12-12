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
	"errors"
	"github.com/jiusanzhou/tentacle/pkg/protocol"
	"os"
	"sync"
)

// These are all the registered plugins.
var (
	// plugins is a map of server type to map of plugin name to
	// Plugin. These are the "general" plugins that may or may
	// not be associated with a specific server type. If it's
	// applicable to multiple server types or the server type is
	// irrelevant, the key is empty string (""). But all plugins
	// must have a name.
	plugins = make(map[string]Plugin)

	// eventHooks is a map of hook name to Hook. All hooks plugins
	// must have a name.
	eventHooks = &sync.Map{}

	// commands is a slice of commands. All commands are store here
	commands = []*Command{}
)

// SetupFunc is used to set up a plugin, or in other words,
// execute a directive. It will be called once per key for
// each server block it appears in.
type SetupFunc func() error

type Plugin struct {
	Setup    SetupFunc
	Commands []interface{}
}

// EventName represents the name of an event used with event hooks.
type EventName string

// Define names for the various events
const (
	StartupEvent         EventName = "startup"
	ShutdownEvent                  = "shutdown"
	CertRenewEvent                 = "certrenew"
	InstanceStartupEvent           = "instancestartup"
	InstanceRestartEvent           = "instancerestart"
)

// OnProcessExit is a list of functions to run when the process
// exits -- they are ONLY for cleanup and should not block,
// return errors, or do anything fancy. They will be run with
// every signal, even if "shutdown callbacks" are not executed.
// This variable must only be modified in the main goroutine
// from init() functions.
var OnProcessExit []func()

// CmdHandler
var NewCmdHandler func(ctx Context) protocol.CmdHandler

// TODO: can we not use interface{}?
func RegisterCommand(hdlr interface{}) {
	cmd, err := NewCommand(hdlr)
	if err != nil {
		panic(err)
	}
	commands = append(commands, cmd)
}

var (
	ErrUnknownCmdType = errors.New("unknown command type")
)

func setup() {

	OnProcessExit = append(OnProcessExit, func() {
		if PidFile != "" {
			os.Remove(PidFile)
		}
	})

	NewCmdHandler = func(ctx Context) protocol.CmdHandler {
		return func(cmdType uint32, data []byte) ([]byte, bool, error) {
			if cmd := commands[cmdType]; cmd != nil {
				res, err := cmd.Exec(ctx, data)
				return res, true, err
			}

			return nil, false, ErrUnknownCmdType
		}
	}
}
