package dns

import (
	"net"
	"testing"
)

func TestQueryRoot(t *testing.T) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		t.Fatalf("network: %s", err)
	}
	defer conn.Close()
	raddr := &net.UDPAddr{net.ParseIP("198.41.0.4"), 53}
	name, err := NewName("liulonnie.net")
	if err != nil {
		t.Fatalf("NewName: %s", err)
	}
	msg := NewQuesMsg(name, A)
	buf, err := msg.Wire()
	if buf == nil || err != nil {
		t.Fatalf("Wire: %s", err)
	}
	n, err := conn.WriteTo(buf, raddr)
	if err != nil {
		t.Fatalf("WriteTo: %s", err)
	}
	t.Logf("sent to server: %d bytes\n", n)
	buf = make([]byte, 512)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		t.Fatalf("recv: %s", err)
	}
	t.Logf("recv from: %s %d bytes\n", addr, n)
	msg, err = FromWire(buf[:n])
	if err != nil {
		t.Fatalf("parse: %s", err)
	}
	s := msg.String()
	t.Logf("msg: \n%s\n", s)
}

func TestQuerier(t *testing.T) {
	name, err := NewName("liulonnie.net")
	if err != nil {
		t.Fatalf("name: %s", err)
	}

	conn, err := NewConn()
	if err != nil {
		t.Fatalf("NewConn: %s", err)
	}

	resp, err := conn.Query(ParseIP("198.41.0.4"), name, A)
	// resp, err := conn.QueryHost(ParseIP("192.168.0.1"), name, A)
	if err != nil {
		t.Fatalf("query: %s", err)
	}
	conn.Close()

	t.Logf("msg: \n%s", resp.Msg)
}
