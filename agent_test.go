package dns

import (
	"bytes"
	"testing"
	// "os"
)

func TestRecurProb(t *testing.T) {
    name := makeName("liulonnie.net")

	conn, err := NewConn()
	if err != nil {
		t.Fatalf("NewConn: %s", err)
	}

	var buf bytes.Buffer
	solver := NewSolver(conn, &buf)
	solver.Solve(NewRecurProb(name, A))

	t.Logf("\n %s", buf.String())

	conn.Close()
}
