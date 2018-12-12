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
	"net"

	"github.com/jiusanzhou/knife-go/config/options"
	"github.com/jiusanzhou/tentacle/tentacle"
	"github.com/soheilhy/cmux"
)

type Server struct {
	instance *tentacle.Instance

	opts   Options
	rawLis net.Listener
	mux    cmux.CMux

	l  net.Listener
	ml cmux.CMux
}

func NewServer(ops ...options.Option) (*Server, error) {
	var s = &Server{
		// set opts from global
		opts: Opts,
	}

	for _, op := range ops {
		op(s.opts)
	}

	l, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		// log.Errorf("listen addr: %s end with error: %s", s.opts.Addr, err)
		return nil, err
	}

	s.ml = cmux.New(l)

	// we chose http2 as tunnel and command conn

	// should we need to use http2???

	return s, nil
}
