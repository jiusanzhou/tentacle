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

package main

import (
	"os"
	"runtime"

	// import all commands we need to load
	// sub commands also need to import at here
	_ "github.com/jiusanzhou/tentacle/cmd/client"
	_ "github.com/jiusanzhou/tentacle/cmd/server"

	// loads plugins from pluings
	_ "github.com/jiusanzhou/tentacle/plugins"

	"github.com/jiusanzhou/tentacle/cmd"
)

func main() {

	// try to set max P
	// maybe this is not necessary
	runtime.GOMAXPROCS(runtime.NumCPU())

	cmd.Run(os.Args[1:])
}
