package proto

import (
	"github.com/jiusanzhou/tentacle/conn"
)

type Protocol interface {
	GetName() string
	WrapConn(conn.Conn, interface{}) conn.Conn
}