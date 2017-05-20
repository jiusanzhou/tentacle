package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/util"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strings"
	"github.com/valyala/fasthttp"
)

type HttpProxyListener struct {
	net.Addr
	Conns chan conn.Conn
}

func httpProxyListen(addr, typ string) (l *HttpProxyListener, err error) {

	// Listen for incoming connections [proxy]
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	l = &HttpProxyListener{
		Addr:  listener.Addr(),
		Conns: make(chan conn.Conn),
	}

	go func() {
		for {
			rawCon, err := listener.Accept()
			if err != nil {
				log.Error("Failed to accept new TCP[HTTP] connection: %v", err)
				continue
			}
			c := conn.Wrap(rawCon, typ)
			c.Info("New connection from %v", c.RemoteAddr())
			l.Conns <- c
		}
	}()

	return
}

func httpListener(addr string, tlsConfig *tls.Config) {
	listener, err := httpProxyListen(addr, "http pxy")
	if err != nil {
		panic(err)
	}

	log.Info("Listening for http proxy on %s", listener.Addr.String())

	// TODO: add username/password supported
	dialer, err := proxy.SOCKS5("tcp", opts.socketAddr, nil, proxy.Direct)
	if err != nil {
		return
	}

	httpTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		},
	}
	// fmt.Println(httpTransport)
	socketProxyClient := &http.Client{Transport: httpTransport}

	for c := range listener.Conns {
		go func(httpConn conn.Conn) {
			// don't crash on panics
			defer func() {
				httpConn.Close()
				if r := recover(); r != nil {
					httpConn.Info("httpHandler failed with error %v: %s", r, debug.Stack())
				}
			}()

			// TODO: change to read []bytes of conn, connect to tunnel directly

			// get http request object
			req, err := http.ReadRequest(bufio.NewReader(httpConn))

			if err != nil {
				httpConn.Warn("read http request error, %v", err)
				return
			}

			// issues:
			// RequestURI is the unmodified Request-URI of the
			// Request-Line (RFC 2616, Section 5.1) as sent by the client
			// to a server. Usually the URL field should be used instead.
			// It is an error to set this field in an HTTP client request.

			// http://stackoverflow.com/questions/19595860/http-request-requesturi-field-when-making-request-in-go
			u, _ := url.Parse(req.RequestURI)
			req.RequestURI = ""
			req.URL = u

			// check username/password
			// TODO: check username/password
			if !auth(req) {
				httpConn.Warn("Auth failed")

				resp := fasthttp.AcquireResponse()
				buf := fasthttp.AcquireByteBuffer()

				resp.SetBody(unauthorizedMsg)
				resp.SetStatusCode(407)
				resp.Header.Set("server", "Tentacle")
				resp.WriteTo(buf)

				// for https
				//if req.Method == "CONNECT" {
				//	httpConn.Write(util.S2b("HTTP/1.0 200 OK\r\n\r\n"))
				//}

				httpConn.Write(buf.B)

				fasthttp.ReleaseResponse(resp)
				fasthttp.ReleaseByteBuffer(buf)

				return
			}

			// https://jdebp.eu./FGA/web-proxy-connection-header.html
			req.Header.Del("Proxy-Connection")
			req.Header.Del("Proxy-Authenticate")
			req.Header.Del("Proxy-Authorization")
			// Connection, Authenticate and Authorization are single hop Header:
			// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
			// 14.10 Connection
			//   The Connection general-header field allows the sender to specify
			//   options that are desired for that particular connection and MUST NOT
			//   be communicated by proxies over further connections.
			req.Header.Del("Connection")

			if req.Method == "CONNECT" {
				// handle https

				host := u.String()

				// TODO: use conenction pool
				// connect to remote with socket proxy
				remoteConn, err := dialer.Dial("tcp", host)

				if err != nil {
					httpConn.Warn("connect to [https]%s error, %v.", host, err)
					return
				}

				wrapedRemoteConn := conn.Wrap(remoteConn, "remote")
				wrapedHttpConn := conn.Wrap(httpConn, "http")

				defer func(){
					wrapedHttpConn.Close()
					wrapedRemoteConn.Close()
				}()

				httpConn.Write(util.S2b("HTTP/1.0 200 Connection Established\r\n\r\n"))

				// copy data
				conn.Join(wrapedHttpConn, wrapedRemoteConn)

			} else {
				// handle http

				// get socket proxy from pool
				// TODO: use connection pool

				// send through socket proxy
				resp, err := socketProxyClient.Do(req)
				if err != nil {
					httpConn.Warn("send request through socket proxy error, %v", err)
					return
				}

				// copy proxy conn and client conn
				rawResp, err := httputil.DumpResponse(resp, true)
				if err != nil {
					httpConn.Warn("dumps response error, %v", err)
					return
				}
				httpConn.Write(rawResp)
			}
		}(c)
	}

}

func handleHttpConn(httpConn conn.Conn) {
	// req := fasthttp.AcquireRequest()
	// buf := fasthttp.AcquireByteBuffer()
	// httpConn.Read(buf.B)
}

var proxyAuthorizationHeader = "Proxy-Authorization"
var unauthorizedMsg = []byte("407 Proxy Authentication Required")

var Authorization string

func auth(req *http.Request) bool {

	if Authorization == "" {
		return true
	}

	authheader := strings.SplitN(req.Header.Get(proxyAuthorizationHeader), " ", 2)
	req.Header.Del(proxyAuthorizationHeader)
	if len(authheader) != 2 || authheader[0] != "Basic" {
		return false
	}

	if authheader[1] != Authorization {
		controlManager.Warn("Should be %s, but is %s", Authorization, authheader[1])
		return false
	} else {
		return true
	}

	//userpassraw, err := base64.StdEncoding.DecodeString(authheader[1])
	//if err != nil {
	//	return false
	//}
	//userpass := strings.SplitN(string(userpassraw), ":", 2)
	//if len(userpass) != 2 {
	//	return false
	//}
	//
	//if Authorization == "" {
	//	return true
	//} else {
	//
	//}
	//
	//return f(userpass[0], userpass[1])
}
