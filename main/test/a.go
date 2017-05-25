package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"os/signal"
	"syscall"
)

type Options struct {
	cmd  string
	addr string
}

var opts Options

func listen(addr string) {
	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatalln("Listen error ", e)
	}

	for {
		conn, e := l.Accept()
		if e != nil {
			fmt.Println("Accept conn error, ", e)
			continue
		}

		go func(conn net.Conn) {
			buf := make([]byte, 1024)
			fmt.Println("Start reading data from ", conn.RemoteAddr().String())
			for {
				n, e := conn.Read(buf)
				if n > 0 {
					fmt.Println("I have read ", n)
				}

				if e != nil {
					fmt.Println("Exit read loop with error, ", e)
					return
				}
			}
		}(conn)
	}
}

func dial(addr string) {
	conn, e := net.Dial("tcp", addr)
	if e!=nil{
		log.Fatalln("Dial error", e)
	}
	buf := []byte("Hello World")
	reap := time.NewTicker(5*time.Second)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-ch:
			conn.Close()
			close(ch)
			reap.Stop()
			fmt.Println("Close this connection.")
			return
		case <-reap.C:
			conn.Write(buf)
		}
	}
}

func init() {
	cmd := flag.String("cmd", "listen", "listen or dial")
	addr := flag.String("addr", ":8080", "address for listen or dial")

	flag.Parse()

	opts = Options{cmd: *cmd, addr: *addr}
}

func main() {
	fmt.Println(opts)
	switch opts.cmd {
	case "listen":
		listen(opts.addr)
	case "dial":
		dial(opts.addr)
	default:
		fmt.Println("Unknow command ", opts.cmd)
		os.Exit(0)
	}
}
