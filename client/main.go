package client

import (
	"fmt"
	"os"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/util"
	"runtime"
	"github.com/inconshreveable/mousetrap"
	"math/rand"

	// for debug pprof
	// _ "net/http/pprof"
	// "net/http"
)

func init() {

	fmt.Println("To honor the memory of fox&rabbit.")

	if runtime.GOOS == "windows" {
		if mousetrap.StartedByExplorer() {
			fmt.Println("You'd better do not double-click tentacler, and I donn't why!")
		}
	}
}

func Main() {

	// for debug pprof
	// go http.ListenAndServe(":8081", http.DefaultServeMux)

	opts, err := ParseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set up logging
	log.LogTo(opts.logto, opts.loglevel)


	// read configuration file
	config, err := LoadConfiguration(opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		fmt.Println("Couldn't securely seed the random number generator!")
		os.Exit(1)
	}

	rand.Seed(seed)

	NewControl(config).Run()
}
