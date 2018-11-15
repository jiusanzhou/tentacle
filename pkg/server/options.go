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
	"github.com/jiusanzhou/tentacle/pkg/options"
)

type Options struct {
	Addr string
	Auth string

	options.Options
}

var (
	opts = Options{
		Addr: DefaultAddr,
		Auth: DefaultAuth,
	}
)

const (
	DefaultAddr    = ":8080"
	DefaultAuth    = "tentacle:123456"
	DefaultTls     = false
	DefaultTimeout = 30
	DefaultMaxConn = 10
)

func init() {
	options.Init(&opts)
	opts.Register("tls", DefaultTls)
	opts.Register("timeout", DefaultTimeout)
	opts.Register("max_conn", DefaultMaxConn)
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
