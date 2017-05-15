package server

import (
	"flag"
	"time"
)

type Options struct {
	controlAddr string
	tunnelAddr  string
	socketAddr  string
	httpAddr    string
	tlsCrt      string
	tlsKey      string
	logto       string
	loglevel    string

	redialInterval time.Duration
}

func parseArgs() *Options {
	controlAddr := flag.String("controlAddr", ":4442", "Public address listening for tentacle client")
	tunnelAddr := flag.String("tunnelAddr", ":4443", "Public address listening for tentacle request")
	socketAddr := flag.String("socketAddr", ":8888", "Public address listening for socket5 proxy")
	httpAddr := flag.String("httpAddr", ":8887", "Public address listening forhttp proxy")
	tlsCrt := flag.String("tlsCrt", "", "Path to a TLS certificate file")
	tlsKey := flag.String("tlsKey", "", "Path to a TLS key file")
	logto := flag.String("log", "stdout", "Write log messages to this file. 'stdout' and 'none' have special meanings")
	loglevel := flag.String("log-level", "DEBUG", "The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	redialInterval := flag.Duration("redial-interval", 1*time.Minute, "Redial interval for each tentacler")
	flag.Parse()

	return &Options{
		controlAddr: *controlAddr,
		tunnelAddr:  *tunnelAddr,
		socketAddr:  *socketAddr,
		httpAddr:    *httpAddr,
		tlsCrt:      *tlsCrt,
		tlsKey:      *tlsKey,
		logto:       *logto,
		loglevel:    *loglevel,

		redialInterval: *redialInterval,
	}
}
