package msg

import (
	"encoding/binary"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/util"
)

func readMsgShared(c conn.Conn) (buf []byte, err error) {
	c.Debug("Waiting to read message")

	//var sz int64
	//err = binary.Read(c, binary.LittleEndian, &sz)
	//if err != nil {
	//	return
	//}
	//c.Debug("Reading message with length: %d", sz)
	//
	//buffer = make([]byte, sz) // ? This may be cause GC
	// use byte pool
	buf = util.GlobalLeakyBuf.Get()
	// CAUTION:
	// remember to put buffer back
	// pack.go L16

	_, err = c.Read(buf)

	// don't check the error
	// what is going on?

	//if int64(n) != sz {
	//	err = errors.New(fmt.Sprintf("Expected to read %d bytes, but only read %d", sz, n))
	//	return
	//}

	return
}

func ReadMsg(c conn.Conn) (msg Message, err error) {
	buffer, err := readMsgShared(c)
	if err != nil {
		return
	}

	return Unpack(buffer)
}

func ReadMsgInto(c conn.Conn, msg Message) (err error) {
	buffer, err := readMsgShared(c)
	if err != nil {
		return
	}
	return UnpackInto(buffer, msg)
}

func WriteMsg(c conn.Conn, msg interface{}) (err error) {
	buffer, err := Pack(msg)
	if err != nil {
		return
	}

	c.Debug("Writing message: %s", string(buffer))
	err = binary.Write(c, binary.LittleEndian, int64(len(buffer)))

	if err != nil {
		return
	}

	if _, err = c.Write(buffer); err != nil {
		return
	}

	return nil
}
