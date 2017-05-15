package server

import (
	"errors"
	"fmt"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/msg"
	"github.com/jiusanzhou/tentacle/util"
	"github.com/jiusanzhou/tentacle/version"
	"io"
	"runtime/debug"
	"time"
)

const (
	pingTimeoutInterval = 30 * time.Second
	connReapInterval    = 10 * time.Second
	controlWriteTimeout = 10 * time.Second
)

type Status struct {
	IsAlive   bool
	Delay     time.Duration
	Bandwidth float32
}

type Control struct {
	// auth message
	auth *msg.Auth

	// actual connection
	conn conn.Conn

	// request id to public connections from user
	connections *util.StringMap

	// ready chan for request
	reqId2ReadyChan *util.StringMap

	// current count of requests through this tunnel
	reqCount int

	// put a message in this channel to send it over
	// conn to the client
	out chan (msg.Message)

	// read from this channel to get the next message sent
	// to us over conn by the client
	in chan (msg.Message)

	// the last time we received a ping from the client - for heartbeats
	lastPing time.Time

	// identifier
	id string

	// synchronizer for controlled shutdown of writer()
	writerShutdown *util.Shutdown

	// synchronizer for controlled shutdown of reader()
	readerShutdown *util.Shutdown

	// synchronizer for controlled shutdown of manager()
	managerShutdown *util.Shutdown

	// synchronizer for controller shutdown of entire Control
	shutdown *util.Shutdown

	// status
	status *Status
}

func (ctl *Control) Write(m msg.Message) {
	ctl.out <- m
}

func (ctl *Control) Read() chan msg.Message {
	return ctl.in
}

func (ctl *Control) GetConn(reqId string) conn.Conn {
	if v, ok := ctl.connections.Get(reqId).(conn.Conn); ok {
		return v
	} else {
		return nil
	}
}

func (ctl *Control) DelConn(reqId string) {
	ctl.connections.Del(reqId)
}

func (ctl *Control) AddConn(reqId string, conn conn.Conn) {
	ctl.connections.Set(reqId, conn)
}

func (ctl *Control) DelReadyChan(reqId string) {
	if c, ok := ctl.reqId2ReadyChan.Get(reqId).(chan struct{}); ok {
		close(c)
		ctl.reqId2ReadyChan.Del(reqId)
	}
}

// not for ready, for connection is done

func (ctl *Control) InitReady(reqId string) {
	ctl.reqId2ReadyChan.Set(reqId, make(chan struct{}))
}

func (ctl *Control) SetReady(reqId string) {
	if c, ok := ctl.reqId2ReadyChan.Get(reqId).(chan struct{}); ok {
		c <- struct{}{}
	} else {
		ctl.conn.Error("No request id, for ready chan")
	}
}

func (ctl *Control) WaitReady(reqId string) error {
	if c, ok := ctl.reqId2ReadyChan.Get(reqId).(chan struct{}); ok {
		<-c
		return nil
	} else {
		return errors.New("No request id, for ready chan")
	}
}

func NewControl(ctlConn conn.Conn, authMsg *msg.Auth) {

	// create the object
	c := &Control{
		auth:            authMsg,
		conn:            ctlConn,
		connections:     util.NewStringMap(),
		reqId2ReadyChan: util.NewStringMap(),
		out:             make(chan msg.Message),
		in:              make(chan msg.Message),
		lastPing:        time.Now(),
		writerShutdown:  util.NewShutdown(),
		readerShutdown:  util.NewShutdown(),
		managerShutdown: util.NewShutdown(),
		shutdown:        util.NewShutdown(),
	}

	failAuth := func(e error) {
		_ = msg.WriteMsg(ctlConn, &msg.AuthResp{Error: e.Error()})
		ctlConn.Close()
	}

	// register the clientid
	c.id = authMsg.ClientId
	if c.id == "" {
		// it's a new session, assign an ID
		c.id = util.RandId(8)
	}

	// set logging prefix
	ctlConn.SetType("ctl")
	ctlConn.AddLogPrefix(c.id)

	if authMsg.Version != version.Proto {
		failAuth(fmt.Errorf("Incompatible versions. Server %s, client %s. Download a new version at http://ngrok.com", version.MajorMinor(), authMsg.Version))
		return
	}

	c.status = &Status{IsAlive: true}

	// register the control
	if replaced := controlManager.AddControl(c.id, c); replaced != nil {
		replaced.shutdown.WaitComplete()
	}

	// start the writer first so that the following messages get sent
	go c.writer()

	// Respond to authentication
	c.out <- &msg.AuthResp{
		Version:    version.Proto,
		MmVersion:  version.MajorMinor(),
		ClientId:   c.id,
		TunnelPort: controlManager.tunnelPort,
	}

	// As a performance optimization, ask for a proxy connection up front
	// c.out <- &msg.ReqProxy{}

	// manage the connection
	go c.manager()
	go c.reader()

	// start redial cmd sending
	go func(c *Control){
		ticker := time.NewTicker(opts.redialInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				func(c *Control){
					defer func(){
						if err := recover(); err != nil {
							return
						}
					}()
					// send redial cmd
					c.out <- &msg.Cmd{
						ClientId: c.id,
						Commands: []string{"tentacler redial"},
					}
				}(c)
			}
		}
	}(c)

	go c.stopper()
}

func (c *Control) manager() {
	// don't crash on panics
	defer func() {
		if err := recover(); err != nil {
			c.conn.Info("Control::manager failed with error %v: %s", err, debug.Stack())
		}
	}()

	// kill everything if the control manager stops
	defer c.shutdown.Begin()

	// notify that manager() has shutdown
	defer c.managerShutdown.Complete()

	// reaping timer for detecting heartbeat failure
	reap := time.NewTicker(connReapInterval)
	defer reap.Stop()

	for {
		select {
		case <-reap.C:
			if time.Since(c.lastPing) > pingTimeoutInterval {
				c.conn.Info("Lost heartbeat")
				c.shutdown.Begin()
			}

		case mRaw, ok := <-c.in:
			// c.in closes to indicate shutdown
			if !ok {
				return
			}

			switch m := mRaw.(type) {
			case *msg.CmdResp:
			case *msg.DialResp:
				// c.SetReady(m.ReqId)
				c.conn.Debug("remote connection of [%s] is ready.", m.ReqId)
			case *msg.Ping:
				c.lastPing = time.Now()
				c.out <- &msg.Pong{}
			}
		}
	}
}

func (c *Control) writer() {
	defer func() {
		if err := recover(); err != nil {
			c.conn.Info("Control::writer failed with error %v: %s", err, debug.Stack())
		}
	}()

	// kill everything if the writer() stops
	defer c.shutdown.Begin()

	// notify that we've flushed all messages
	defer c.writerShutdown.Complete()

	// write messages to the control channel
	for m := range c.out {
		c.conn.SetWriteDeadline(time.Now().Add(controlWriteTimeout))
		if err := msg.WriteMsg(c.conn, m); err != nil {
			panic(err)
		}
	}
}

func (c *Control) reader() {
	defer func() {
		if err := recover(); err != nil {
			c.conn.Warn("Control::reader failed with error %v: %s", err, debug.Stack())
		}
	}()

	// kill everything if the reader stops
	defer c.shutdown.Begin()

	// notify that we're done
	defer c.readerShutdown.Complete()

	// read messages from the control channel
	for {
		if msg, err := msg.ReadMsg(c.conn); err != nil {
			if err == io.EOF {
				c.conn.Info("EOF")
				return
			} else {
				panic(err)
			}
		} else {
			// this can also panic during shutdown
			c.in <- msg
		}
	}
}

func (c *Control) stopper() {
	defer func() {
		if r := recover(); r != nil {
			c.conn.Error("Failed to shut down control: %v", r)
		}
	}()

	// wait until we're instructed to shutdown
	c.shutdown.WaitBegin()

	// remove self from the control registry
	controlManager.DelControl(c.id)

	// shutdown manager() so that we have no more work to do
	close(c.in)
	c.managerShutdown.WaitComplete()

	// shutdown writer()
	close(c.out)
	c.writerShutdown.WaitComplete()

	// close connection fully
	c.conn.Close()

	c.shutdown.Complete()
	c.conn.Info("Shutdown complete")
}

// Called when this control is replaced by another control
// this can happen if the network drops out and the client reconnects
// before the old tunnel has lost its heartbeat
func (c *Control) Replaced(replacement *Control) {
	c.conn.Info("Replaced by control: %s", replacement.conn.Id())

	// set the control id to empty string so that when stopper()
	// calls registry.Del it won't delete the replacement
	c.id = ""

	// tell the old one to shutdown
	c.shutdown.Begin()
}

func (c *Control) Id() string {
	return c.id
}

func (c *Control)SetDeath() {
	c.status.IsAlive = false
	controlManager.Resort()
}
