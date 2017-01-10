package client

import (
	"fmt"
	"os"
	"github.com/jiusanzhou/tentacle/log"
)

func init() {}

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set up logging
	log.LogTo(opts.logto, opts.loglevel)

	fmt.Print(opts)
}
