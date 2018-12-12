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
	"fmt"
	"testing"
)

type hello struct {
	name  string
	exits bool
	count int
}

func (h *hello) Marshal() ([]byte, error) {
	return []byte(h.name), nil
}

func (h *hello) Unmarshal(data []byte) error {
	h.name = string(data)
	return nil
}

func (h *hello) fn(ctx Context, hx *hello) (*hello, error) {
	fmt.Println(hx.name)
	hx.name = "AAAAAAAAAAAAAAAA"
	return hx, nil
}

func TestNewCommand(t *testing.T) {
	cmd0, err := NewCommand((&hello{name:"A"}).fn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cmd0.Exec(Context{}, []byte("SSS")))
	fmt.Println(cmd0.Exec(Context{}, []byte("SSS")))

	cmd1, _ := NewCommand(func(ctx Context, h *hello) error {
		fmt.Println(h.name)
		return nil
	})
	fmt.Println(cmd1.Exec(Context{}, []byte("AAA")))
}
