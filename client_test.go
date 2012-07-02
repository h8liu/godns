package dns

import (
	"bytes"
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient()

	// query
	resp, err := c.Query(ParseIP("198.41.0.4"), makeName("liulonnie.net"), A)
	if err != nil {
		t.Fatalf("query: %s", err)
	}
	t.Logf("resp: \n%s", resp.Msg)

	// solve
	var buf bytes.Buffer
	p := NewRecurProb(makeName("liulonnie.net"), A)
	c.Solve(p, &buf)
	t.Logf("\n%s", buf.String())
}
