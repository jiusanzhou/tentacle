package server

import (
	"encoding/binary"
	"errors"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/msg"
	"github.com/jiusanzhou/tentacle/util"
	"io"
	"net"
	"runtime/debug"
	"strconv"
	"time"
	"crypto/tls"
	"github.com/jiusanzhou/tentacle/conn"
)

var (
	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support 1 method now")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")
)

const (
	socksVer5       = 5
	socksCmdConnect = 1
)

const (
	idVer   = 0
	idCmd   = 1
	idType  = 3 // address type index
	idIP0   = 4 // ip addres start index
	idDmLen = 4 // domain address length index
	idDm0   = 5 // domain address start index

	typeIPv4 = 1 // type is ipv4 address
	typeDm   = 3 // type is domain address
	typeIPv6 = 4 // type is ipv6 address

	lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
	lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
	lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
)

type SocketProxyListener struct {
	net.Addr
	Conns chan conn.Conn
}

// Listen for incoming proxy
func socketListener(addr string, tlsConfig *tls.Config) {
	// listen for incoming connections
	listener, err := socketProxyListen(addr, "socket pxy")
	if err != nil {
		panic(err)
	}

	log.Info("Listening for socket proxy on %s", listener.Addr.String())

	for c := range listener.Conns {
		go func(socketConn conn.Conn) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					socketConn.Info("socketHandler failed with error %v: %s", r, debug.Stack())
				}
			}()

			if err = HandShake(socketConn); err != nil {
				log.Error("socks handshake:", err)
				// may be a http request
				return
			}

			rawAddr, addr, err := GetRequest(socketConn)
			if err != nil {
				log.Error("error getting request:", err)
				return
			}

			// Sending connection established message immediately to client.
			// This some round trip time for creating socks connection with the client.
			// But if connection failed, the client will get connection reset error.
			_, err = socketConn.Write([]byte{
				0x05, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00})

			if err != nil {
				log.Debug("send connection confirmation:", err)
				return
			}

			// save conn with reqId to map
			reqId := util.RandId(8)
			controlManager.AddConn(reqId, socketConn)

			// get a controller to proxy this request
			ctl := controlManager.GetControlByRequestId(reqId)
			if ctl == nil {
				// socketConn.Write(util.S2b(BadGateway))
				log.Error("Cann't Get control tunnel.")
				return
			}

			// add this req conn to ctl
			// ctl.AddConn(reqId, socketConn)
			// defer ctl.DelConn(reqId)

			// init ready chan
			// ctl.InitReady(reqId)
			// defer ctl.DelReadyChan(reqId)

			ctl.Write(&msg.Dial{
				ClientId: ctl.Id(),
				ReqId:    reqId,
				RawAddr:  rawAddr,
				Addr:     addr,
			})

			// wait for remote is ready
			// why should wait?
			// wait to close connections

			// ctl.WaitReady(reqId)

			// ctl.DelReadyChan(reqId)
			// controlManager.DelConn(reqId)
			// socketConn.Close()

			// TODO: shut down the remote connection

		}(c)
	}
}

func socketProxyListen(addr, typ string) (l *SocketProxyListener, err error) {

	// Listen for incoming connections [proxy]
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	l = &SocketProxyListener{
		Addr:  listener.Addr(),
		Conns: make(chan conn.Conn),
	}

	go func() {
		for {
			rawCon, err := listener.Accept()
			if err != nil {
				log.Error("Failed to accept new TCP connection: %v", err)
				continue
			}

			c := conn.Wrap(rawCon, typ)
			c.Info("New connection from %v", c.RemoteAddr())

			l.Conns <- c
		}
	}()

	return
}

func HandShake(conn conn.Conn) (err error) {
	const (
		idVer     = 0
		idNmethod = 1
	)
	// version identification and method selection message in theory can have
	// at most 256 methods, plus version and nmethod field in total 258 bytes
	// the current rfc defines only 3 authentication methods (plus 2 reserved),
	// so it won't be such long in practice

	buf := make([]byte, 258)

	var n int

	conn.SetDeadline(time.Now().Add(connReadTimeout))

	// make sure we get the nmethod field
	if n, err = io.ReadAtLeast(conn, buf, idNmethod+1); err != nil {
		return
	}

	if buf[idVer] != socksVer5 {
		return errVer
	}

	nmethod := int(buf[idNmethod])
	msgLen := nmethod + 2

	if n == msgLen {
		// handshake done, common case
		// do nothing, jump directly to send confirmation
	} else if n < msgLen { // has more methods to read, rare case
		if _, err = io.ReadFull(conn, buf[n:msgLen]); err != nil {
			return
		}
	} else {
		// error, should not get extra data
		return errAuthExtraData
	}

	// send confirmation: version 5, no authentication required
	_, err = conn.Write([]byte{socksVer5, 0})
	return
}

func GetRequest(conn conn.Conn) (rawaddr []byte, host string, err error) {

	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)

	var n int
	conn.SetDeadline(time.Now().Add(connReadTimeout))

	// read till we get possible domain length field
	if n, err = io.ReadAtLeast(conn, buf, idDmLen+1); err != nil {
		return
	}
	// check version and cmd
	if buf[idVer] != socksVer5 {
		err = errVer
		return
	}

	if buf[idCmd] != socksCmdConnect {
		err = errCmd
		return
	}

	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
	default:
		err = errAddrType
		return
	}

	if n == reqLen {
		// common case, do nothing
	} else if n < reqLen { // rare case
		if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return
		}
	} else {
		err = errReqExtraData
		return
	}

	rawaddr = buf[idType:reqLen]

	switch buf[idType] {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}

	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))

	return
}
