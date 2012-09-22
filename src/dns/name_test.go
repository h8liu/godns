package dns

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

	n = N(".")
	if !n.IsRoot() {
		t.Error(". is not root")
	}

	n = N("google.com")
	if !n.Parent().Equal(N("com")) {
		t.Error("parent of google.com is not com")
	}

	// TODO: need more test cases
}
