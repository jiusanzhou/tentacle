package msg

import (
	"github.com/pquerna/ffjson/ffjson"
	"errors"
	"fmt"
	"reflect"
)

func unpack(buffer []byte, msgIn Message) (msg Message, err error) {
	var env Envelope
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
	return ffjson.Marshal(struct {
		Type    string
		Payload interface{}
	}{
		Type:    reflect.TypeOf(payload).Elem().Name(),
		Payload: payload,
	})
}