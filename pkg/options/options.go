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
	"errors"
	"reflect"
)

type Key string

type Value struct {
	typ          reflect.Type
	value        reflect.Value
	defaultValue reflect.Value
}

// Options ...
type Options map[Key]*Value

// Add function dynamic add a options arg for opts
// only called this at setup
func (opts Options) Add(key string, value interface{}) error {
	_, ok := opts[Key(key)]
	if ok {
		return errors.New("duplicate key")
	}
	opts[Key(key)] = &Value{
		typ:          reflect.TypeOf(value),
		value:        nil,
		defaultValue: reflect.ValueOf(value),
	}
	return nil
}

func (opts Options) Set(key string, value interface{}) error {
	if v, ok := opts[Key(key)]; ok {
		if v.typ == reflect.TypeOf(value) {
			v.value = reflect.ValueOf(value)
			return nil
		}
		return errors.New("value with key type is not correct")
	}
	return errors.New("no suck key of options")
}

func (opts Options) Get(key string) reflect.Value {
	if v := opts[Key(key)]; v != nil {
		if v.value.IsNil() {
			return v.defaultValue
		}
		return v.value
	}
	panic("no such key of options")
}

// NewOptions implements a new Options
func NewOptions() Options {
	return make(Options)
}
