package msg

import (
	"errors"
	"fmt"
	"reflect"
	// "gopkg.in/vmihailenco/msgpack.v2"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/jiusanzhou/tentacle/util"
)

func unpack(buffer []byte, msgIn Message) (msg Message, err error) {
	var env Envelope

	// MUST release buffer
	defer util.GlobalLeakyBuf.Put(buffer)

	if err = ffjson.Unmarshal(buffer, &env); err != nil {
		return
	}

	if msgIn == nil {
		t, ok := TypeMap[env.Type]

		if !ok {
			err = errors.New(fmt.Sprintf("Unsupported message type %s", env.Type))
			return
		}

		// guess type
		msg = reflect.New(t).Interface().(Message)
	} else {
		msg = msgIn
	}

	err = ffjson.Unmarshal(env.Payload, &msg)
	return
}

func UnpackInto(buffer []byte, msg Message) (err error) {
	_, err = unpack(buffer, msg)
	return
}

func Unpack(buffer []byte) (msg Message, err error) {
	return unpack(buffer, nil)
}

func Pack(payload interface{}) ([]byte, error) {
	return ffjson.Marshal(&struct {
		Type    string
		Payload interface{}
	}{
		Type:    reflect.TypeOf(payload).Elem().Name(),
		Payload: payload,
	})
}
