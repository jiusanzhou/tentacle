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


// These are all the registered plugins.
var (
	// serverTypes is a map of registered server types.
	serverTypes = make(map[string]ServerType)

	// plugins is a map of server type to map of plugin name to
	// Plugin. These are the "general" plugins that may or may
	// not be associated with a specific server type. If it's
	// applicable to multiple server types or the server type is
	// irrelevant, the key is empty string (""). But all plugins
	// must have a name.
	plugins = make(map[string]map[string]Plugin)
)

// SetupFunc is used to set up a plugin, or in other words,
// execute a directive. It will be called once per key for
// each server block it appears in.
type SetupFunc func() error

type Plugin interface {
	ServerType() ServerType
	SetupFunc()
}

// ServerType contains information about a server type.
type ServerType struct {
	// Function that returns the list of directives, in
	// execution order, that are valid for this server
	// type. Directives should be one word if possible
	// and lower-cased.
	Directives func() []string
}