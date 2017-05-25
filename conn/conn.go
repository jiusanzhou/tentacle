package conn

import (
	"crypto/tls"
	"fmt"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/util"
	"github.com/valyala/fasthttp"
	"io"
	"math/rand"
	"net"
	"sync"
)

// Connection for controlling and data transportation.
type Conn interface {
	net.Conn
	log.Logger
	Id() string
	SetType(string)
	CloseRead() error
}

type loggedConn struct {
	tcp *net.TCPConn
	net.Conn
	log.Logger
	id  int32
	typ string
}

type Listener struct {
	net.Addr
	Conns chan *loggedConn
}

func wrapConn(conn net.Conn, typ string) *loggedConn {
	switch c := conn.(type) {
	case *loggedConn:
		return c
	case *net.TCPConn:
		wrapped := &loggedConn{c, conn, log.NewPrefixLogger(), rand.Int31(), typ}
		wrapped.AddLogPrefix(wrapped.Id())
		return wrapped
	default:
		return nil
	}
}

func Listen(addr, typ string, tlsCfg *tls.Config) (l *Listener, err error) {
	// Listen for incoming connections
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	l = &Listener{
		Addr:  listener.Addr(),
		Conns: make(chan *loggedConn),
	}

	go func(l *Listener) {
		for {
			rawCon, err := listener.Accept()
			if err != nil {
				log.Error("Failed to accept new TCP connection: %v", err)
				continue
			}

			c := wrapConn(rawCon, typ)
			if tlsCfg != nil {
				c.Conn = tls.Server(c.Conn, tlsCfg)
			}
			c.Info("New connection from %v", c.RemoteAddr())

			l.Conns <- c
		}
	}(l)
	return
}

func Wrap(conn net.Conn, typ string) *loggedConn {
	return wrapConn(conn, typ)
}

func Dial(addr, typ string, tlsCfg *tls.Config) (conn *loggedConn, err error) {

	// Dial to the target address
	var rawConn net.Conn
	if rawConn, err = fasthttp.Dial(addr); err != nil {
		return
	}

	conn = wrapConn(rawConn, typ)
	conn.Debug("New connection to: %v", rawConn.RemoteAddr())

	if tlsCfg != nil {
		conn.StartTLS(tlsCfg)
	}

	return
}

func (c *loggedConn) StartTLS(tlsCfg *tls.Config) {
	c.Conn = tls.Client(c.Conn, tlsCfg)
}

func (c *loggedConn) Close() (err error) {
	if err := c.Conn.Close(); err == nil {
		c.Debug("Closing")
	}
	return
}

func (c *loggedConn) Id() string {
	return fmt.Sprintf("[%x]", c.id)
}

func (c *loggedConn) CloseRead() error {
	// XXX: use CloseRead() in Conn.Join() and in Control.shutdown() for cleaner
	// connection termination. Unfortunately, when I've tried that, I've observed
	// failures where the connection was closed *before* flushing its write buffer,
	// set with SetLinger() set properly (which it is by default).
	return c.tcp.CloseRead()
}

func (c *loggedConn) SetType(typ string) {
	oldId := c.Id()
	c.typ = typ
	c.ClearLogPrefixes()
	c.AddLogPrefix(c.Id())
	c.Info("Renamed connection %s", oldId)
}

func pipe(to Conn, from Conn, bytesCopied *int64, wait *sync.WaitGroup) {
	buf := util.GlobalLeakyBuf.Get()
	defer func() {
		util.GlobalLeakyBuf.Put(buf)
		to.Close()
		from.Close()
		wait.Done()
	}()

	var err error
	// *bytesCopied, err = io.Copy(to, from)
	*bytesCopied, err = io.CopyBuffer(to, from, buf)
	if err != nil {
		from.Warn("Copied %d bytes to %s before failing with error %v", *bytesCopied, to.Id(), err)
	} else {
		from.Debug("Copied %d bytes to %s", *bytesCopied, to.Id())
	}
}

func PipeConn(to, from Conn) {
	buf := util.GlobalLeakyBuf.Get()
	defer func() {
		util.GlobalLeakyBuf.Put(buf)
		to.Close()
		from.Close()
		// wait.Done()
	}()

	for {
		n, err := from.Read(buf)

		// read may return EOF with n > 0
		// should always process n > 0 bytes before handling error
		if n > 0 {
			// Note: avoid overwrite err returned by Read.
			if _, err := to.Write(buf[0:n]); err != nil {
				from.Warn("Copied %d bytes to %s before failing with error %v", n, to.Id(), err)
				break
			}
		}

		if err != nil {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			break
		}
	}
}

func Join(c Conn, c2 Conn) (int64, int64) {
	var wait sync.WaitGroup

	wait.Add(2)
	var fromBytes, toBytes int64
	go pipe(c, c2, &fromBytes, &wait)
	go pipe(c2, c, &toBytes, &wait)
	c.Info("Joined with connection %s", c2.Id())
	wait.Wait()
	return fromBytes, toBytes
}
