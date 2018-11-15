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

package options

import (
	"fmt"
	"testing"
	"time"
)

type MyOptions struct {
	Max    int64
	Addr   string
	needMe bool
	Options
}

var opts = &MyOptions{}

func TestInit(t *testing.T) {
	Init(opts)
	if opts.Options == nil {
		t.Fatal("init failed")
	}
}

func TestOptions_Set(t *testing.T) {
	Init(opts)
	opts.Set("Addr", "HHHH")
	opts.Set("Max", int64(6))
	fmt.Println(opts)
	fmt.Println(opts.Addr)
	opts.Register("newAddr", ":8080")
	fmt.Println(opts.Get("Addr").String())
	fmt.Println(opts.Get("newAddr").String())
}

func TestOptions_Register(t *testing.T) {
	Init(opts)
	opts.Register("a", 1)
	opts.Register("b", 1*time.Second)
	opts.Register("c", "C")
	opts.Register("d", []string{"d"})
	opts.Register("e", true)
	opts.Register("f", 2.1)

	fmt.Println(opts.Get("f").Float())
	fmt.Println(opts.Get("e").Bool())
	fmt.Println(opts.Get("d").Interface().([]string))
	fmt.Println(opts.Get("c").String())
	fmt.Println(opts.Get("b").Interface().(time.Duration))
	fmt.Println(opts.Get("a").Int())
}