package dns

import (
	"bytes"
	"testing"
	// "os"
)

func TestRecur(t *testing.T) {
	name := makeName("liulonnie.net")

	conn, err := NewConn()
	if err != nil {
		t.Fatalf("NewConn: %s", err)
	}

	var buf bytes.Buffer
	solver := NewSolver(conn, &buf)
	solver.Solve(NewRecurProb(name, A).Prob())

	t.Logf("\n %s", buf.String())

	conn.Close()
}
