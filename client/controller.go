package client

import (
	"fmt"
	"github.com/jiusanzhou/tentacle/client/mvc"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/msg"
	"github.com/jiusanzhou/tentacle/util"
	"github.com/jiusanzhou/tentacle/version"
	"github.com/valyala/fasthttp"
	"math"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	defaultServerAddr = "0.0.0.0:4442"
	pingInterval      = 20 * time.Second
	maxPongLatency    = 15 * time.Second
)

type command interface{}

type cmdQuit struct {
	// display this message after quit
	message string
}

type Control struct {
	// Controller logger
	log.Logger

	// the model sends updates through this broadcast channel
	// updates *util.Broadcast

	// internal structure to issue commands to the controller
	cmds chan command

	id string

	// server address
	serverAddr string
	serverHost string

	tunnelAddr string

	ctlConn *conn.Conn

	connStatus mvc.ConnStatus

	authToken string

	tunnelPool conn.Pool

	remoteConns *util.StringMap

	tunnelConns *util.StringMap

	serverVersion string

	// options
	config *Configuration

	configPath string

	GetTunnelConn func() conn.Conn
}

func (ctl *Control) GetConn(k string) conn.Conn {
	if c, ok := ctl.remoteConns.Get(k).(conn.Conn); ok {
		return c
	} else {
		return nil
	}
}

func (ctl *Control) SetConn(k string, conn conn.Conn) {
	ctl.remoteConns.Set(k, conn)
}

func (ctl *Control) DelConn(k string) {
	ctl.remoteConns.Del(k)
}

func (ctl *Control) GetTunnel(k string) conn.Conn {
	if c, ok := ctl.tunnelConns.Get(k).(conn.Conn); ok {
		return c
	} else {
		return nil
	}
}

func (ctl *Control) SetTunnel(k string, conn conn.Conn) {
	ctl.tunnelConns.Set(k, conn)
}

func (ctl *Control) DelTunnel(k string) {
	ctl.tunnelConns.Del(k)
}

func (ctl *Control) getTunnelDirect() conn.Conn {
	tunnelRawConn, err := fasthttp.Dial(ctl.tunnelAddr)
	if err != nil {
		return nil
	}
	tunnelConn := conn.Wrap(tunnelRawConn, "tunnel")
	tunnelConn.SetDeadline(time.Time{})
	return tunnelConn
}

func (ctl *Control) getTunnelFromPool() conn.Conn {

	var err error
	if ctl.tunnelPool == nil {
		ctl.Error("Tunnel pool is nil")
		return ctl.getTunnelDirect()
	}

	conn, err := ctl.tunnelPool.Get()
	if err != nil {
		ctl.Error("Get connection from tunnel pool error, %v", err)
		return ctl.getTunnelDirect()
	}

	return conn
}

// Hearbeating to ensure our connection tentacled is still live
func (c *Control) heartbeat(lastPongAddr *int64, conn conn.Conn) {
	lastPing := time.Unix(atomic.LoadInt64(lastPongAddr)-1, 0)
	ping := time.NewTicker(pingInterval)
	pongCheck := time.NewTicker(time.Second)

	defer func() {
		conn.Close()
		ping.Stop()
		pongCheck.Stop()
	}()

	for {
		select {
		case <-pongCheck.C:
			lastPong := time.Unix(0, atomic.LoadInt64(lastPongAddr))
			needPong := lastPong.Sub(lastPing) < 0
			pongLatency := time.Since(lastPing)

			if needPong && pongLatency > maxPongLatency {
				c.Info("Last ping: %v, Last pong: %v", lastPing, lastPong)
				c.Info("Connection stale, haven't gotten PongMsg in %d seconds", int(pongLatency.Seconds()))
				return
			}

		case <-ping.C:
			err := msg.WriteMsg(conn, &msg.Ping{})
			if err != nil {
				conn.Debug("Got error %v when writing PingMsg", err)
				return
			}
			lastPing = time.Now()
		}
	}
}

// Establishes and manages a tunnel control connection with the server
func (ctl *Control) control() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("control recovering from failure %v", r)
		}
	}()

	// establish control channel
	var (
		ctlConn conn.Conn
		err     error
	)

	ctlConn, err = conn.Dial(ctl.serverAddr, "ctl", nil)

	ctl.ctlConn = &ctlConn

	if err != nil {
		panic(err)
	}

	defer ctlConn.Close()

	// authenticate with the server
	auth := &msg.Auth{
		ClientId:  ctl.id,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   version.Proto,
		MmVersion: version.MajorMinor(),
		User:      ctl.authToken,
	}

	if err = msg.WriteMsg(ctlConn, auth); err != nil {
		panic(err)
	}

	// wait for the server to authenticate us
	var authResp msg.AuthResp
	if err = msg.ReadMsgInto(ctlConn, &authResp); err != nil {
		panic(err)
	}

	if authResp.Error != "" {
		emsg := fmt.Sprintf("Failed to authenticate to server: %s", authResp.Error)
		ctl.Shutdown(emsg)
		return
	}

	ctl.tunnelAddr = fmt.Sprintf("%s:%s", ctl.serverHost, authResp.TunnelPort)

	if ctl.config.PoolSize == 0 {
		ctl.GetTunnelConn = ctl.getTunnelDirect
	}else {
		ctl.initTunnelPool()
	}

	ctl.id = authResp.ClientId
	SaveCacheId(ctl.id)
	ctl.serverVersion = authResp.MmVersion
	ctl.Info("Authenticated with server, client id: %v", ctl.id)

	if err = SaveAuthToken(ctl.configPath, ctl.authToken); err != nil {
		ctl.Error("Failed to save auth token: %v", err)
	}

	// start the heartbeat
	lastPong := time.Now().UnixNano()
	ctl.Go(func() { ctl.heartbeat(&lastPong, ctlConn) })

	ctl.connStatus = mvc.ConnOnline
	ctl.Info("Connectted to tantacle service %v ", ctl.serverAddr)

	// main control loop
	for {
		var rawMsg msg.Message
		if rawMsg, err = msg.ReadMsg(ctlConn); err != nil {
			panic(err)
		}

		switch m := rawMsg.(type) {
		case *msg.Pong:
			ctl.Go(func() { atomic.StoreInt64(&lastPong, time.Now().UnixNano()) })
		case *msg.Cmd:
			ctl.Go(func() { ctl.handleCmd(m) })
		case *msg.Dial:
			ctl.Go(func() { ctl.handleDial(m) })
		default:
			ctlConn.Warn("Ignoring unknown control message %v ", m)
		}
	}
}

func (ctl *Control) handleCmd(m *msg.Cmd) {
	ctl.Debug("Handle command from [%s]", m.ClientId)
	stdOut := []string{}
	for _, c := range m.Commands {
		o, e := util.DoCommand(c)
		ctl.Debug("Run command: %s", c)
		if e == nil {
			stdOut = append(stdOut, util.B2s(o))
		} else {
			stdOut = append(stdOut, "Command error: "+e.Error())
		}
	}
	ctl.Info("Command stdout: %v", stdOut)
}

func (ctl *Control) handleDial(m *msg.Dial) {

	// TODO: use connections' pool
	// dial to the remote
	remoteRawConn, err := fasthttp.Dial(m.Addr)
	if err != nil {
		ctl.Error("Dial to remote %s error, %v", m.Addr, err)
		return
	}
	remoteConn := conn.Wrap(remoteRawConn, "remote")
	remoteConn.SetDeadline(time.Time{})
	defer remoteConn.Close()

	// get tunnel connection
	tunnelConn := ctl.GetTunnelConn()
	if tunnelConn == nil {
		ctl.Warn("Get tunnel conn error.")
		return
	}
	defer tunnelConn.Close()

	// send regTun msg back though this tunnel conn
	msg.WriteMsg(tunnelConn, &msg.RegTun{
		ClientId: ctl.id,
		ReqId:    m.ReqId,
	})

	// if has data from dial msg
	// we should send it to remote immediately
	if len(m.Data) > 0 {
		_, err := remoteConn.Write(m.Data)
		if err != nil {
			ctl.Error("Send data to remote error, %v", err)
			return
		}
	}

	// pipe copy data through those two connections
	// conn.Join(tunnelConn, remoteConn)
	// BUG: wait for a long time
	conn.Join(remoteConn, tunnelConn)

	// remove remote and tunnel
	// if use tunnel pool, release this tunnel
	// ctl.remoteConns.Del(m.ReqId)
	// ctl.tunnelConns.Del(m.ReqId)

}

func (ctl *Control) Run() {

	// how long we should wait before we reconnect
	maxWait := 30 * time.Second
	wait := 1 * time.Second

	ctl.Go(func() {
		for {
			// run the control channel
			ctl.control()

			// control only returns when a failure has occurred, so we're going to try to reconnect
			if ctl.connStatus == mvc.ConnOnline {
				wait = 1 * time.Second
			}

			log.Info("Waiting %d seconds before reconnecting", int(wait.Seconds()))
			time.Sleep(wait)
			// exponentially increase wait time
			wait = 2 * wait
			wait = time.Duration(math.Min(float64(wait), float64(maxWait)))
			ctl.connStatus = mvc.ConnReconnecting
		}
	})

	done := make(chan int)
	for {
		select {
		case obj := <-ctl.cmds:
			switch cmd := obj.(type) {
			case cmdQuit:
				msg := cmd.message
				go func() {
					ctl.Info(msg)
					for _, k := range ctl.remoteConns.GetKeys() {
						if c := ctl.remoteConns.Get(k).(conn.Conn); c != nil {
							c.Close()
						}
					}
					done <- 1
				}()
			}

		case <-done:
			return
		}
	}
}

func (ctl *Control) Shutdown(message string) {
	ctl.cmds <- cmdQuit{message: message}
}

func (ctl *Control) Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := util.MakePanicTrace(r)
				ctl.Error(err)
				ctl.Shutdown(err)
			}
		}()

		fn()
	}()
}

// public interface
func NewControl(config *Configuration) *Control {
	ctl := &Control{
		Logger: log.NewPrefixLogger("controller"),
		// updates: util.NewBroadcast(),

		// server address
		serverAddr: config.ServerAddr,

		// connection status
		connStatus: mvc.ConnConnecting,

		// auth token
		authToken: config.AuthToken,

		cmds: make(chan command),

		remoteConns: util.NewStringMap(),

		tunnelConns: util.NewStringMap(),

		config: config,

		// config path
		configPath: config.Path,
	}

	host, _, err := net.SplitHostPort(ctl.serverAddr)
	if err != nil {
		panic("server address error.")
	}
	ctl.serverHost = host

	ctl.id = GetCachedId()

	return ctl
}


func (ctl *Control) initTunnelPool(){
	var err error
	max := ctl.config.PoolSize
	min := int(math.Ceil(float64(max / 5)))
	ctl.tunnelPool, err = conn.NewChannelPool(min, max,
		func() (conn.Conn, error) {
			tunnelRawConn, err := fasthttp.Dial(ctl.tunnelAddr)
			if err != nil {
				return nil, err
			}
			tunnelConn := conn.Wrap(tunnelRawConn, "tunnel")
			tunnelConn.SetDeadline(time.Time{})
			return tunnelConn, nil
		})
	if err != nil {
		ctl.Error("Init tunnel pool error, %v", err)
		ctl.GetTunnelConn = ctl.getTunnelDirect
	} else {
		ctl.Info("Init tunnel pool with min %d, max %d", min, max)
		ctl.GetTunnelConn = ctl.getTunnelFromPool
	}
}