package dns2

import (
    "testing"
    "net"
)

func TestQueryRoot(t *testing.T) {
    conn, err := net.ListenPacket("udp4", ":0")
    if err != nil { t.Fatalf("network: %s", err) }
    defer conn.Close()
    raddr := &net.UDPAddr{net.ParseIP("198.41.0.4"), 53}
    msg, err:= QuesMsg(".", NS)
    if err != nil { t.Fatalf("QuesMsg: %s", err) }
    buf, err := msg.ToRaw()
    if err != nil { t.Fatalf("ToRaw: %s", err) }
    buf = make([]byte, 512)
    n, addr, err := conn.ReadFrom(buf)
    if err != nil { t.Fatalf("recv: %s", err) }
    fmt.Logf("recv size: %d", n)
}
