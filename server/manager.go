package server

import (
	"errors"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/msg"
	"github.com/jiusanzhou/tentacle/util"
	"math/rand"
	"time"
)

// Manager maps a client ID to Control structures
type Manager struct {
	resort chan struct{}

	// client id to control tunnel of client
	controls *util.StringMap

	controlIds []string

	ctlCount int

	// request id to connection from public connections
	connections *util.StringMap

	// ready chan for request
	reqId2ReadyChan *util.StringMap

	// relationship from request id to client id
	reqId2CliId *util.StringMap

	tunnelPort string

	log.Logger
}

func NewManager() *Manager {

	rand.Seed(time.Now().UTC().UnixNano())

	m := &Manager{
		resort:          make(chan struct{}),
		controls:        util.NewStringMap(),
		controlIds:      []string{},
		connections:     util.NewStringMap(),
		reqId2ReadyChan: util.NewStringMap(),
		reqId2CliId:     util.NewStringMap(),
		Logger:          log.NewPrefixLogger("manager", "ctl"),
	}

	go m.ResortControls()
	go m.redialManager()

	return m
}

// not for ready, for connection is done
// if over time not ready
// clear it

func (m *Manager) DelReadyChan(reqId string) {
	if c, ok := m.reqId2ReadyChan.Get(reqId).(chan struct{}); ok {
		close(c)
		m.reqId2ReadyChan.Del(reqId)
	}
}

func (m *Manager) InitReady(reqId string) {
	m.reqId2ReadyChan.Set(reqId, make(chan struct{}))
}

func (m *Manager) SetReady(reqId string) {
	if c, ok := m.reqId2ReadyChan.Get(reqId).(chan struct{}); ok {
		c <- struct{}{}
	} else {
		m.Error("No request id, for ready chan")
	}
}

func (m *Manager) WaitReady(reqId string, timeout time.Duration) error {
	if c, ok := m.reqId2ReadyChan.Get(reqId).(chan struct{}); ok {
		timer := time.NewTimer(timeout)
		select {
		case <-timer.C:
			return errors.New("Time out")
		case <-c:
			timer.Stop()
			return nil
		}
	}

	return nil
}

// for controller's tunnel

func (m *Manager) AddControl(clientId string, ctl *Control) (oldCtl *Control) {
	oldCtlV := m.controls.Get(clientId)
	if oldCtlV != nil {
		// m.ctlCount--
		oldCtl = oldCtlV.(*Control)
		oldCtl.Replaced(ctl)
	}
	// m.ctlCount++
	m.Resort()
	m.controls.Set(clientId, ctl)
	m.Info("Registered control tunnel with id %s", clientId)
	return
}

func (m *Manager) GetControl(clientId string) *Control {
	if c, ok := m.controls.Get(clientId).(*Control); ok {
		return c
	} else {
		return nil
	}
}

func (m *Manager) DelControl(clientId string) {
	v := m.controls.Del(clientId)
	if v {
		// m.ctlCount--
		m.Resort()
		m.Info("Remove control with id %s", clientId)
	} else {
		m.Info("Conn't find control with id %s", clientId)
	}
}

func (m *Manager) RandGetControl() *Control {

	// a way is directly use map's RandGet function
	// this is not a good way

	//if c, ok := m.controls.RandGet().(*Control); ok {
	//	return c
	//} else {
	//	return nil
	//}

	// more smart method is using a goroutin to sort keys
	// we get a key randomly
	if m.ctlCount == 0 {
		return nil
	}
	return m.GetControl(m.controlIds[rand.Intn(m.ctlCount)])

}

// main method to get control by request id
// if request is not in reqId2cliId, will get a new controller
// and set the cliId to reqId

func (m *Manager) GetControlByRequestId(requestId string) *Control {
	if clientId, ok := m.reqId2CliId.Get(requestId).(string); ok {
		return m.GetControl(clientId)
	} else {
		if ctl := m.RandGetControl(); ctl != nil {
			ctl.reqCount++
			m.reqId2CliId.Set(requestId, ctl.id)
			return ctl
		} else {
			return nil
		}
	}
}

// for request's connection

func (m *Manager) AddConn(requestId string, conn conn.Conn) {
	m.connections.Set(requestId, conn)
	m.InitReady(requestId)
}

func (m *Manager) GetConn(requestId string) conn.Conn {
	if c, ok := m.connections.Get(requestId).(conn.Conn); ok {
		return c
	} else {
		return nil
	}
}

func (m *Manager) DelConn(requestId string) {
	m.connections.Del(requestId)

	// get ctl id
	if ctlId, ok := m.reqId2CliId.Get(requestId).(string); ok && ctlId != "" {
		if ctl := m.GetControl(ctlId); ctl != nil {
			ctl.reqCount--
		}
	}

	m.reqId2CliId.Del(requestId)
	m.DelReadyChan(requestId)
}

func (m *Manager) Resort() {
	m.resort <- struct{}{}
}

func (m *Manager) ResortControls() {
	for {
		// we should make that
		// if we change(add or del) the controls
		// or update status of controller
		// must m.resort <- struct{}{}
		<-m.resort
		m.sortControls()
	}
}

func (m *Manager) sortControls() {
	keys := m.controls.GetKeys()
	ids := []string{}

	for _, k := range keys {

		// get ctl out
		c := m.GetControl(k)
		if c == nil {
			continue
		}

		// check is alive
		if c.status.Code == StatusCodeAlive {
			ids = append(ids, k)
		}

		// TODO: check other meta info and sort by them
	}

	m.controlIds = ids
	m.ctlCount = len(m.controlIds)
}

// redial manager

func (m *Manager) redialManager() {

	// redial one by one
	ticker := time.NewTicker(opts.redialInterval)
	defer func() {
		// recover()
		ticker.Stop()

	}()
	index := 0
	for {
		<-ticker.C
		// if has error break out
		// send redial cmd

		if index >= m.ctlCount {
			if m.ctlCount == 0 {
				continue
			}
			index = 0
		}
		if ctlId := m.controlIds[index]; ctlId != "" {
			m.Debug("redial controller [%s]", ctlId)
			m.redialController(ctlId)
			index++
		} else {
			m.Debug("can not get control id from: [%d]", index)
		}
	}

}

func (m *Manager) redialController(ctlId string) {
	// send redial signal to controller
	if ctl := m.GetControl(ctlId); ctl != nil {
		if ctl.reqCount <= 0 {
			m.Debug("set %s preparing for redial", ctlId)
			// set ctl to preparing
			ctl.SetPreparing()

			// send redial signal
			ctl.out <- &msg.Cmd{
				ClientId: ctl.id,
				Commands: []string{"tentacler redial"},
			}
		} else {
			// set ctl to wait preparing
			m.Debug("set %s waiting for preparing for redial", ctlId)
			ctl.SetWaitPreparing()
			go func(ctl *Control) {
				// check if req count equals zero
				ticker := time.NewTicker(200 * time.Millisecond)
			loop:
				for {
					<-ticker.C
					if ctl.reqCount <= 0 {
						m.Debug("set %s preparing for redial, break loop", ctlId)
						// set ctl to preparing
						ctl.SetPreparing()

						// send redial signal
						ctl.out <- &msg.Cmd{
							ClientId: ctl.id,
							Commands: []string{"tentacler redial"},
						}
						break loop
					}
				}
			}(ctl)
		}
	}
}
