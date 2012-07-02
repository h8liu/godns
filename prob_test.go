package dns

import (
	"bytes"
	"testing"
	// "os"
)

func TestRecur(t *testing.T) {
	name := makeName("liulonnie.net")

	conn := NewConn()
	var buf bytes.Buffer
	solver := NewSolver(conn, &buf)
	solver.Solve(NewRecurProb(name, A))

	t.Logf("\n%s", buf.String())

	conn.Close()
}
