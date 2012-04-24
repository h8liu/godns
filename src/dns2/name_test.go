package dns2

import (
	"testing"
)

func TestName(t *testing.T) {
	var n *Name

	n, e := NewName("")
	if !(n == nil && e.Error() == "name '': empty name") {
		t.Error("should be empty name error")
	}

	n, e = NewName("a.b.")
	if e != nil || n == nil {
		t.Error("error on valid name 'a.b.'")
	}
	if n.String() != "a.b" {
		t.Error("final dot not truncated on 'a.b.'")
	}

	// TODO: need more test cases
}
