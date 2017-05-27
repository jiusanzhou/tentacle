package server

import (
	"bufio"
	"crypto/tls"
	"github.com/jiusanzhou/tentacle/conn"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/msg"
	"github.com/jiusanzhou/tentacle/util"
	"github.com/valyala/fasthttp"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strings"
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

	for c := range listener.Conns {
		go func(httpConn conn.Conn) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					httpConn.Close()
					httpConn.Info("httpHandler failed with error %v: %s", r, debug.Stack())
				}
			}()

			// TODO: change to read []bytes of conn, connect to tunnel directly

			// get http request object
			req, err := http.ReadRequest(bufio.NewReader(httpConn))

			if err != nil {
				httpConn.Warn("read http request error, %v", err)
				writeResponse(httpConn, serverError, 500)
				httpConn.Close()
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
				writeResponse(httpConn, unauthorizedMsg, 407)
				httpConn.Close()
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

			// save conn with reqId to map
			reqId := util.RandId(8)
			controlManager.AddConn(reqId, httpConn)

			// get a controller to proxy this request
			ctl := controlManager.GetControlByRequestId(reqId)
			if ctl == nil {
				// socketConn.Write(util.S2b(BadGateway))
				httpConn.Error("Cann't Get control tunnel.")
				writeResponse(httpConn, serverError, 500)
				httpConn.Close()
				controlManager.DelConn(reqId)
				return
			}

			if req.Method == "CONNECT" {

				// handle https

				ctl.Write(&msg.Dial{
					ClientId: ctl.Id(),
					ReqId:    reqId,
					Addr:     u.String(),
				})

				httpConn.Write(util.S2b("HTTP/1.1 200 Connection Established\r\n\r\n"))
			} else {

				// handle http
				rawReq, err := httputil.DumpRequestOut(req, true)

				if err != nil {
					httpConn.Error("Dump request error, %v.", err)
					writeResponse(httpConn, proxyServerMsg, 200)
					httpConn.Close()
					controlManager.DelConn(reqId)
					return
				}

				// send the request to tunnel, how to do this?

				host := u.Host
				if !strings.Contains(host, ":") {
					host = host + ":80"
				}

				ctl.Write(&msg.Dial{
					ClientId: ctl.Id(),
					ReqId:    reqId,
					Addr:     host,
					Data:     rawReq,
				})

				// we must close this http conn if we finished one request.
			}

			// wait for ready
			err = controlManager.WaitReady(reqId, readyTimeout)
			if err!=nil{
				httpConn.Error("Dial request timeout")
				httpConn.Close()
				controlManager.DelConn(reqId)
			}
		}(c)
	}

}

var proxyAuthorizationHeader = "Proxy-Authorization"
var unauthorizedMsg = []byte("407 Proxy Authentication Required")
var proxyServerMsg = []byte("This is a proxy server")
var serverError = []byte("Server error, please contact admin ---<a href=\"mailto:jsz3@live.com\">Zoe</a>")

func writeResponse(c conn.Conn, b []byte, code int) {

	resp := fasthttp.AcquireResponse()
	buf := fasthttp.AcquireByteBuffer()

	resp.SetBody(b)
	resp.SetStatusCode(code)
	resp.Header.Set("server", "Tentacle")
	resp.WriteTo(buf)

	c.Write(buf.B)

	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseByteBuffer(buf)
}

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
