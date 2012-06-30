package dns

import (
	"bytes"
	"testing"
	// "os"
)

func TestAgent(t *testing.T) {
	name, err := NewName("liulonnie.net")
	if err != nil {
		t.Fatalf("name: %s", err)
	}
	conn, err := NewConn()
	if err != nil {
		t.Fatalf("NewConn: %s", err)
	}

	agent := NewAgent(conn, nil)
	resp := agent.Query(ParseIP("198.41.0.4"), name, A)
	if resp != nil {
		t.Logf("msg: %s", resp.Msg)
	} else {
		t.Logf("unreachable")
	}
	conn.Close()

}

func TestRecurProb(t *testing.T) {
	name, err := NewName("liulonnie.net")
	if err != nil {
		t.Fatalf("name: %s", err)
	}
	conn, err := NewConn()

	if err != nil {
		t.Fatalf("NewConn: %s", err)
	}

	var buf bytes.Buffer
	agent := NewAgent(conn, &buf)
	root, err := NewName(".")
	if err != nil {
		t.Fatalf("name: %s", err)
	}
	ips := []*IPv4{ParseIP("198.41.0.4")}
	servers := []*NameServer{
		&NameServer{name: root, ips: ips},
	}
	prob := NewRecurProb(name, A)
	prob.StartFrom(root, servers)
	agent.Solve(prob)
	agent.FlushLog()

	t.Logf("\n %s", buf.String())

	conn.Close()
}
