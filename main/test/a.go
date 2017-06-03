package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	if e != nil {
		log.Fatalln("Dial error", e)
	}
	buf := []byte("Hello World")
	reap := time.NewTicker(5 * time.Second)
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

type Handler struct{}

func removeProxyHeaders(r *http.Request) {
	r.RequestURI = "" // this must be reset when serving a request with the client
	// If no Accept-Encoding header exists, Transport will add the headers it can accept
	// and would wrap the response body with the relevant reader.
	r.Header.Del("Accept-Encoding")
	// curl can add that, see
	// https://jdebp.eu./FGA/web-proxy-connection-header.html
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	// Connection, Authenticate and Authorization are single hop Header:
	// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
	// 14.10 Connection
	//   The Connection general-header field allows the sender to specify
	//   options that are desired for that particular connection and MUST NOT
	//   be communicated by proxies over further connections.
	r.Header.Del("Connection")
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {

	} else {
		removeProxyHeaders(r)
		b, _ := httputil.DumpRequest(r, true)
		r.Body.Close()
		fmt.Println(string(b))
	}
}

func Main() {
	h := &Handler{}
	http.ListenAndServe(":8080", h)
}
