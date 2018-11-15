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

const (
	fieldName = "Options"
	fieldType = "options.Options"
)

type vKey string

type vValue struct {
	typ          reflect.Type
	value        reflect.Value
	defaultValue reflect.Value
}

type Option func(opts Options)

// Options ...
type Options interface {
	Register(key string, defaultValue interface{})
	Set(key string, value interface{})
	Get(key string) reflect.Value
	Copy(v interface{})
}

type options struct {
	typ     reflect.Type
	val     reflect.Value
	fields  map[vKey]*vValue
	extends map[vKey]*vValue
}

// Register ...
//
// panic if with error
func (opts *options) Register(key string, defaultValue interface{}) {
	if opts.extends == nil {
		Init(opts)
	}
	if err := opts.Add(key, defaultValue); err != nil {
		panic(err)
	}
}

// Add function dynamic add a options arg for opts
// only called this at setup
func (opts *options) Add(key string, value interface{}) error {
	// try to check fields
	if _, ok := opts.fields[vKey(key)]; ok {
		panic("duplicate key with struct fields")
	}

	if _, ok := opts.extends[vKey(key)]; ok {
		panic("duplicate key")
	}
	opts.extends[vKey(key)] = &vValue{
		typ: reflect.TypeOf(value),
		// value:        nil,
		defaultValue: reflect.ValueOf(value),
	}
	return nil
}

// Set ...
func (opts *options) Set(key string, value interface{}) {
	var err error

	// first we try to set value of filed of struct
	if v, ok := opts.fields[vKey(key)]; ok {
		// panic("duplicated field name with struct")
		// opts.val.FieldByName(key).Set(reflect.ValueOf(value))
		// check the type
		if v.typ != reflect.TypeOf(value) {
			panic("value with key type is not correct")
		}
		v.value.Set(reflect.ValueOf(value))
		return
	}

	if v, ok := opts.extends[vKey(key)]; ok {
		if v.typ == reflect.TypeOf(value) {
			v.value = reflect.ValueOf(value)
			return
		}
		err = errors.New("value with key type is not correct")
	}
	err = errors.New("no suck key of options")

	if err != nil {
		panic(err)
	}
}

// Get ...
func (opts *options) Get(key string) reflect.Value {

	// try to get value from struct
	if v, ok := opts.fields[vKey(key)]; ok {
		// return opts.val.FieldByName(key)
		return v.value
	}

	if v := opts.extends[vKey(key)]; v != nil {
		if !v.value.IsValid() {
			return v.defaultValue
		}
		return v.value
	}

	panic("no such key of options")
}

// Copy ...
// can we have a better way?
func (opts *options) Copy(v interface{}) {
	// TODO: deep copy
}

// Init ...
//
// init
func Init(opts interface{}) {
	var typ = reflect.TypeOf(opts)
	var val = reflect.ValueOf(opts)
	if typ.Kind() != reflect.Ptr {
		panic("you must offer a addressed struct as input")
	}
	typ = typ.Elem()
	val = val.Elem()
	var opt, ok = typ.FieldByName(fieldName)

	if !ok {
		panic("you must offer a struct with type options.Options without named ")
	}

	if fieldType != opt.Type.String() {
		panic("Field of struct Options must be type of options.Options")
	}

	var o = &options{
		typ,
		val,
		make(map[vKey]*vValue),
		make(map[vKey]*vValue),
	}

	val.FieldByName(fieldName).Set(reflect.ValueOf(o))

	// should we need to loads all fields?
	var total = typ.NumField() - 1
	for i := 0; i < total; i++ {
		var fs = typ.Field(i)
		if fs.Name != fieldName {
			// try to set
			o.fields[vKey(fs.Name)] = &vValue{
				typ:   fs.Type,
				value: val.FieldByName(fs.Name),
			}
		}
	}
}
