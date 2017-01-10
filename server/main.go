package server

import (
	"fmt"
	"github.com/jiusanzhou/tentacle/version"
)

func Main() {
	fmt.Println("Tentacle version: ", version.Full())
}
