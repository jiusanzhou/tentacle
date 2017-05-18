package server

import (
	"crypto/tls"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/msg"
	"github.com/jiusanzhou/tentacle/util"
	"math/rand"
	"net"
	"runtime/debug"
	"time"
	// for debug pprof
	// _ "net/http/pprof"
	// "net/http"
	"encoding/base64"
)

const (
	registryCacheSize uint64        = 1024 * 1024 // 1 MB
	connReadTimeout   time.Duration = 10 * time.Second
)

// GLOBALS
var (
	controlManager *Manager

	// XXX: kill these global variables - they're only used in tunnel.go for constructing forwarding URLs
	opts *Options
	// listeners map[string]*conn.Listener
)

func Main() {

	// parse options
	opts = parseArgs()

	// init http proxy author
	if opts.password != "" || opts.username != "" {
		Authorization = base64.StdEncoding.EncodeToString(util.S2b(opts.username + ":" + opts.password))
	}

	// init logging
	log.LogTo(opts.logto, opts.loglevel)

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		panic(err)
	}
	rand.Seed(seed)

	// init control manager
	controlManager = NewManager()

	// start listeners
	// listeners = make(map[string]*conn.Listener)

	var tlsConfig *tls.Config

	_, port, err := net.SplitHostPort(opts.tunnelAddr)
	if err != nil {
		panic("tunnel address error.")
	}

	controlManager.tunnelPort = port

	// for debug pprof
	// go http.ListenAndServe(":8080", http.DefaultServeMux)

	// listen for socket proxy
	go socketListener(opts.socketAddr, tlsConfig)

	// listen for http proxy
	go httpListener(opts.httpAddr, tlsConfig)

	// listen for tunnel
	go tunnelListener(opts.tunnelAddr, tlsConfig)

	// listen for tentacle clients
	controlListener(opts.controlAddr, tlsConfig)
}

// Listen for incoming control connections, ctl and data with the same tunnel
// We listen for incoming control and proxy connections on the same port
// for ease of deployment. The hope is that by running on port 443, using
// TLS and running all connections over the same port.
func controlListener(addr string, tlsConfig *tls.Config) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "ctl", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control connection on %s", listener.Addr.String())
	for c := range listener.Conns {
		go func(controlConn conn.Conn) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					controlConn.Info("controlListener failed with error %v: %s", r, debug.Stack())
				}
			}()

			controlConn.SetReadDeadline(time.Now().Add(connReadTimeout))
			var rawMsg msg.Message
			if rawMsg, err = msg.ReadMsg(controlConn); err != nil {
				controlConn.Warn("Failed to read message: %v", err)
				controlConn.Close()
				return
			}

			// don't timeout after the initial read, tunnel heart beating will kill
			// dead connections
			controlConn.SetReadDeadline(time.Time{})

			switch m := rawMsg.(type) {
			case *msg.Auth:
				NewControl(controlConn, m)
			default:
				controlConn.Close()
			}
		}(c)
	}
}
