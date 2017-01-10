package client

import (
	"flag"
	"fmt"
	"os"
	"github.com/jiusanzhou/tentacle/version"
)

const usage string = `
Advanced usage: tentacler [OPTIONS] <command> [command args] [...]
	tentacler info				List info from tentacled service.
	tentacler start [tcp] [...]		Start and regist to tentacled service.
	tentacler help				Print help
	tentacler version			Print tentacle version

Examples:
	tentacler start ilife codertool
	tentacler -log=stdout -config=tentacler.yml start
	tentacler version
`

type Options struct {
	config   string
	logto    string
	loglevel string
	command  string
	args     []string
}

func ParseArgs() (opts *Options, err error) {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, usage)
	}

	config := flag.String(
		"config",
		"",
		"Path to ngrok configuration file. (default: $HOME/.tentacler)")

	logto := flag.String(
		"log",
		"none",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	loglevel := flag.String(
		"log-level",
		"DEBUG",
		"The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	flag.Parse()

	opts = &Options{
		config:    *config,
		logto:     *logto,
		loglevel:  *loglevel,
		command:   flag.Arg(0),
	}

	switch opts.command {
	case "info":
		opts.args = flag.Args()[1:]
	case "start":
		opts.args = flag.Args()[1:]
	case "version":
		fmt.Println(version.MajorMinor())
		os.Exit(0)
	case "help":
		flag.Usage()
		os.Exit(0)
	default:
		if len(flag.Args()) > 1 {
			err = fmt.Errorf("You may only specify one port to tunnel to on the command line, got %d: %v",
				len(flag.Args()),
				flag.Args())
			return
		}

		opts.command = "default"
		opts.args = flag.Args()
	}

	return
}
