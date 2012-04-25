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
	msg, err := QuesMsg("liulonnie.net", A)
	if msg == nil || err != nil {
		t.Fatalf("QuesMsg: %s", err)
	}
	buf, err := msg.ToWire()
	if buf == nil || err != nil {
		t.Fatalf("ToWire: %s", err)
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
	t.Logf("msg: \n %s\n", s)
}
