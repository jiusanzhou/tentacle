package client

import (
	"net"
	"github.com/jiusanzhou/tentacle/log"
)

type RemoteConn struct {
	// actual connection
	tcp *net.TCPConn

	net.Conn
	log.Logger
	id  string
	typ string
}

func (s *RemoteConn) Id() string {
	return s.id
}

func (s *RemoteConn) SetType(typ string) {
	s.typ = typ
}

func (s *RemoteConn) CloseRead() error {
	return s.tcp.CloseRead()

}

func wrapRemoteConn(conn net.Conn, id string) *RemoteConn {
	switch c := conn.(type) {
	case *RemoteConn:
		return c
	case *net.TCPConn:
		return &RemoteConn{c, conn, log.NewPrefixLogger(), id, "remote"}
	default:
		return nil
	}
}
