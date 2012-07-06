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
	rp := NewRecurProb(makeName("liulonnie.net"), A)
	c.Solve(rp, &buf)
	t.Logf("\n%s", buf.String())

	buf.Reset()
	ap := NewAddrProb(makeName("liulonnie2.net"))
	c.Solve(ap, &buf)
	t.Logf("\n%s", buf.String())
	for _, ip := range ap.Ips {
		t.Logf("%s\n", ip.String())
	}
}
