package conn

import (
	"testing"
	"github.com/jiusanzhou/eagle/util"
	"fmt"
)

var (
	addrs = []string{":6666", ":6667"}
)

func listen(addr string) (*Listener, error) {
	return Listen(addr, "test", nil)
}

func dial(addr string) (Conn, error) {
	return Dial(addr, "test", nil)
}

func TestListenAndDial(t *testing.T) {
	for _, addr := range addrs {
		l, e := listen(addr)
		if e != nil {
			t.Errorf("Listen %s error, %v", addr, e)
		}
		go func(l *Listener) {}(l)
	}
}

func TestJoin(t *testing.T) {
	l, e := listen(addrs[0])
	if e != nil {
		t.Errorf("Listen %s error, %v", addrs[0], e)
	}

	conn2 := []Conn{}

	go func(){
		for c := range l.Conns {
			conn2 = append(conn2, c)
			if len(conn2) == 2{
				go Join(conn2[0], conn2[1])
				break
			}
		}
	}()

	var c1, c2 Conn

	c1, e = dial(addrs[0])
	if e != nil {
		t.Errorf("Dial %s error, %v", addrs[0], e)
	}

	c2, e = dial(addrs[0])
	if e != nil {
		t.Errorf("Dial %s error, %v", addrs[0], e)
	}

	b := util.S2b("Hello World!")

	c1.Write(b)

	buf := make([]byte, len(b))
	c2.Read(buf)

	for i, v := range b {
		if v != buf[i] {
			t.Errorf("Should get %v on %d, but get %v.", v, i, buf[i])
		}
	}

	c1.Close()
	c2.Close()

	for _, c := range conn2 {
		c.Close()
	}
}
