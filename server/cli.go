package server

import (
	"flag"
	"fmt"
	"github.com/jiusanzhou/tentacle/version"
	"os"
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

	dialTimeout time.Duration

	username string
	password string
}

func parseArgs() *Options {
	controlAddr := flag.String("controlAddr", ":4442", "Public address listening for tentacle client")
	tunnelAddr := flag.String("tunnelAddr", ":4443", "Public address listening for tentacle request")
	socketAddr := flag.String("socketAddr", ":8888", "Public address listening for socket5 proxy")
	httpAddr := flag.String("httpAddr", ":8887", "Public address listening forhttp proxy")
	tlsCrt := flag.String("tlsCrt", "", "Path to a TLS certificate file")
	tlsKey := flag.String("tlsKey", "", "Path to a TLS key file")
	logto := flag.String("log", "stdout", "Write log messages to this file. 'stdout' and 'none' have special meanings")
	loglevel := flag.String("log-level", "INFO", "The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	user := flag.String("user", "", "Http proxy username")
	pass := flag.String("pass", "", "Http proxy password")

	redialInterval := flag.Duration("redial-interval", 1*time.Minute, "Redial interval for each tentacler")

	dialTimeout := flag.Duration("dialTimeout", 2*time.Second, "Timeout for dialing remote, only for http proxy")

	flag.Parse()

	if len(flag.Args()) > 0 && flag.Args()[0] == "version" {
		fmt.Println(version.Full())
		os.Exit(0)
	}

	return &Options{
		controlAddr: *controlAddr,
		tunnelAddr:  *tunnelAddr,
		socketAddr:  *socketAddr,
		httpAddr:    *httpAddr,
		tlsCrt:      *tlsCrt,
		tlsKey:      *tlsKey,
		logto:       *logto,
		loglevel:    *loglevel,

		username: *user,
		password: *pass,

		redialInterval: *redialInterval,

		dialTimeout: *dialTimeout,
	}
}
