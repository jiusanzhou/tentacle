package server

import (
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/util"
	"math/rand"
	"time"
	"io"
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

	// relationship from request id to client id
	reqId2CliId *util.StringMap

	tunnelPort string

	log.Logger
}

func NewManager() *Manager {

	rand.Seed(time.Now().UTC().UnixNano())

	m := &Manager{
		resort:      make(chan struct{}),
		controls:    util.NewStringMap(),
		controlIds:  []string{},
		connections: util.NewStringMap(),
		reqId2CliId: util.NewStringMap(),
		Logger:      log.NewPrefixLogger("manager", "ctl"),
	}

	go m.ResortControls()

	return m
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
	m.reqId2CliId.Del(requestId)
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
		if c.status.IsAlive {
			ids = append(ids, k)
		}

		// TODO: check other meta info and sort by them
	}

	m.controlIds = ids
	m.ctlCount = len(m.controlIds)
}