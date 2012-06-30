package dns

import (
    "testing"
)

func TestAgent(t *testing.T) {
    name, err := NewName("liulonnie.net")
    if err != nil { t.Fatalf("name: %s", err) }

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
}
