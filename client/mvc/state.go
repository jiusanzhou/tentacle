package mvc


import (
	metrics "github.com/rcrowley/go-metrics"
)

type UpdateStatus int

const (
	UpdateNone = -1 * iota
	UpdateInstalling
	UpdateReady
	UpdateAvailable
)

type ConnStatus int

const (
	ConnConnecting = iota
	ConnReconnecting
	ConnOnline
)

type Tunnel struct {
	ClientId string
	ReqId string
	RawAddr string
}

type ConnectionContext struct {
	Tunnel     Tunnel
	ClientAddr string
}

type State interface {
	GetClientVersion() string
	GetServerVersion() string
	GetUpdateStatus() UpdateStatus
	GetConnStatus() ConnStatus
	GetConnectionMetrics() (metrics.Meter, metrics.Timer)
	GetBytesInMetrics() (metrics.Counter, metrics.Histogram)
	GetBytesOutMetrics() (metrics.Counter, metrics.Histogram)
	SetUpdateStatus(UpdateStatus)
}
