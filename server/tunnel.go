package server

import (
	"crypto/tls"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/msg"
	"net"
	"runtime/debug"
	"time"
)

/**
 * Tunnel: A control connection, metadata and proxy connections which
 *         transport data.
 */
type Tunnel struct {
	// request that opened the tunnel
	// req *msg.ReqTunnel

	// time when the tunnel was opened
	start time.Time

	// public url
	url string

	// tcp listener
	listener *net.TCPListener

	// control connection
	ctl *Control

	// logger
	log.Logger

	// closing
	closing int32
}

func NewTunnel(tunnelConn conn.Conn, regTunMsg *msg.RegTun) {

	var clientConn conn.Conn

	// CAUTION: never close tunnel connection
	defer func() {
		tunnelConn.Close()
	}()

	// first should set the remote connected

	// we have do this before
	// tunnelConn.SetDeadline(time.Time{})

	// get control

	// get public conn
	clientConn = controlManager.GetConn(regTunMsg.ReqId)

	clientConn.SetDeadline(time.Time{})

	if clientConn == nil {
		tunnelConn.Error("get client connection error.")
		return
	}
	defer clientConn.Close()

	// I can not close connection with http://loudong.360.cn/help/plan
	// use a fake way to close

	// pipe copy data from public and tunnel
	// BUG: wait for a long time
	conn.Join(tunnelConn, clientConn)
	// conn.Join(clientConn, tunnelConn)

	controlManager.DelConn(regTunMsg.ReqId)

	//if c := controlManager.GetControl(regTunMsg.ClientId); c != nil {
	//	c.SetReady(regTunMsg.ReqId)
	//}
}

func tunnelListener(addr string, tlsConfig *tls.Config) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "tun", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for tunnel connection on %s", listener.Addr.String())
	for c := range listener.Conns {
		go func(tunnelConn conn.Conn) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					tunnelConn.Info("tunnelListener failed with error %v: %s", r, debug.Stack())
				}
			}()

			// don't timeout after the initial read, tunnel heart beating will kill
			// dead connections
			tunnelConn.SetReadDeadline(time.Now().Add(connReadTimeout))
			var rawMsg msg.Message
			if rawMsg, err = msg.ReadMsg(tunnelConn); err != nil {
				tunnelConn.Warn("Failed to read message: %v", err)
				tunnelConn.Close()
				return
			}

			// don't timeout after the initial read, tunnel heart beating will kill
			// dead connections
			tunnelConn.SetDeadline(time.Time{})

			switch m := rawMsg.(type) {
			case *msg.RegTun:
				controlManager.SetReady(m.ReqId)
				NewTunnel(tunnelConn, m)
			default:
				tunnelConn.Close()
			}
		}(c)
	}

}
