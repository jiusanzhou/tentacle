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
	"errors"
	"reflect"
)

var (
	ErrUnspportedFuncSign      = errors.New("command handler must be a function signed: func (context.Content, Marshaler) (Marshaler, error)")
	ErrArgMustBeContext        = errors.New("command handler's first must be context.Content")
	ErrUnsupportedCmdHandler   = errors.New("command handler must be a function")
	ErrUnsupportedLengthOfArgs = errors.New("command handler must signed with 2 arguments.")
	ErrArgMustBeUnmarshaler    = errors.New("command handler's arg must be unmarshaler")
	ErrUnsupportedReturns      = errors.New("command handler's returns must be 2 or 1")
	ErrUnsupportedSigned       = errors.New("command handler's returns must signed: (Marshaler, error) or (error)")
)

// Marshaler is the interface implemented by types that
// can marshal themselves into
type Marshaler interface {
	Marshal() ([]byte, error)
}

// Unmarshaler is the interface implemented by types
// that can unmarshal
type Unmarshaler interface {
	Unmarshal([]byte) error
}

type fn struct {
	origin interface{}
	typ    reflect.Type
	val    reflect.Value
}

type request struct {
	origin Unmarshaler
	typ    reflect.Type
	val    reflect.Value
}

type response struct {
	origin Marshaler
	typ    reflect.Type
	val    reflect.Value
}

type Command struct {
	fn       fn
	request  request
	response response
	returns  []reflect.Value
	cancel   bool

	Name string
}

func (cmd *Command) Exec(ctx Context, data []byte) ([]byte, error) {

	var err error
	var resp []byte

	// get struct out from data
	// TODO: do we need to do reflect.New?
	req, ok := reflect.New(cmd.request.typ).Interface().(Unmarshaler)
	if !ok {
		return nil, ErrArgMustBeUnmarshaler
	}
	err = req.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	// call the command handler
	returns := cmd.fn.val.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(req),
	})

	var errIndex = 1

	if cmd.cancel {
		errIndex = 0
	} else {
		mayResp := returns[0].Interface()
		if mayResp != nil {
			resp, err = (mayResp.(Marshaler)).Marshal()
		}
	}

	mayErr := returns[errIndex].Interface()
	if mayErr != nil {
		err = mayErr.(error)
	}

	return resp, err
}

// NewCommand returns Command:
// fn must be a function witch has 2 arguments and 1 or
// 2 returns.
// If returns' length is 2, means we need to
// send data next
// Length of returns must be 1 or 2
// arguments: 0
// 0: context
// 1: MarshalAble
// returns: [0], 1
// 0: MarshalAble
// 1: error
// Example: func(Request) (Response, error)
//
// ...
func NewCommand(hdl interface{}) (cmd *Command, err error) {

	var ok bool

	// cmd.origin = fn
	cmd = &Command{}

	fnTyp := reflect.TypeOf(hdl)
	fnVal := reflect.ValueOf(hdl)

	cmd.fn = fn{
		origin: hdl,
		typ:    fnTyp,
		val:    fnVal,
	}

	// check typ of fn, must be function
	if fnTyp.Kind() != reflect.Func {
		err = ErrUnsupportedCmdHandler
		return
	}

	// check inputs of fn, must be 2
	if fnTyp.NumIn() != 2 {
		err = ErrUnsupportedLengthOfArgs
		return
	}

	// make sure first argument must be context
	if fnTyp.In(0) != reflect.TypeOf((*Context)(nil)).Elem() {
		err = ErrArgMustBeContext
		return
	}

	// make sure input must can be Marhasler and Unmarhsaler
	cmd.request = request{
		typ: fnTyp.In(1).Elem(),
	}

	cmd.request.origin, ok = reflect.New(cmd.request.typ).Interface().(Unmarshaler)

	if !ok {
		err = ErrArgMustBeUnmarshaler
		return
	}

	var errTyp reflect.Type

	switch fnTyp.NumOut() {
	case 2:
		cmd.response = response{
			typ: fnTyp.Out(0).Elem(),
		}
		cmd.response.origin, ok = reflect.New(cmd.response.typ).Interface().(Marshaler)
		if !ok {
			err = ErrUnsupportedSigned
			return
		}

		errTyp = fnTyp.Out(1)
	case 1:
		cmd.cancel = true
		errTyp = fnTyp.Out(0)
	default:
		err = ErrUnsupportedReturns
		return
	}

	if errTyp != reflect.TypeOf((*error)(nil)).Elem() {
		err = ErrUnsupportedSigned
		return
	}
	// set cmd name from function
	// cmd.Name = runtime.FuncForPC(fnVal.Pointer()).Name()

	return
}
