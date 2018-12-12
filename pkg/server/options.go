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

package server

import (
	"github.com/jiusanzhou/knife-go/config/options"
)

type Options struct {
	Addr string
	Auth string

	options.Options
}

var (
	Opts = Options{
		Addr: DefaultAddr,
		Auth: DefaultAuth,
	}
)

const (
	DefaultAddr    = ":8080"
	DefaultAuth    = "tentacle:123456"
	DefaultTLS     = false
	DefaultTimeout = 30
	DefaultMaxConn = 10
)

func init() {
	options.Init(&Opts)
	Opts.Register("tls", DefaultTLS)
	Opts.Register("timeout", DefaultTimeout)
	Opts.Register("max_conn", DefaultMaxConn)
}

var (
	Addr = func(addr string) options.Option {
		return func(opts options.Options) {
			opts.Set("Addr", addr)
		}
	}

	Auth = func(auth string) options.Option {
		return func(opts options.Options) {
			opts.Set("Auth", auth)
		}
	}

	Tls = func(tls bool) options.Option {
		return func(opts options.Options) {
			opts.Set("tls", tls)
		}
	}

	Timeout = func(timeout int) options.Option {
		return func(opts options.Options) {
			opts.Set("timeout", timeout)
		}
	}

	MaxConn = func(max_conn int) options.Option {
		return func(opts options.Options) {
			opts.Set("max_conn", max_conn)
		}
	}
)
