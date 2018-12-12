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

package conn

import "github.com/jiusanzhou/knife-go/config/options"

type Options struct {
	options.Options
}

var (
	Opts = Options{}
)

func init() {
	options.Init(&Opts)
	Opts.Register("proxy", "")
	Opts.Register("tls", false)
	Opts.Register("timeout", 30)
	Opts.Register("max_conn", 10)
	Opts.Register("over_http", false)
}

var (
	Proxy = func(proxy string) options.Option {
		return func(opts options.Options) {
			opts.Set("proxy", proxy)
		}
	}

	Timeout = func(timeout int) options.Option {
		return func(opts options.Options) {
			opts.Set("timeout", timeout)
		}
	}

	MaxConn = func(maxConn int) options.Option {
		return func(opts options.Options) {
			opts.Set("max_conn", maxConn)
		}
	}
)
