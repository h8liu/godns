package dns

import (
	"testing"
)

func testToken(t *testing.T, s string, e string) {
	if Tokenize(s) != e {
		t.Errorf("tokenize: %s   expect: %s", s, e)
	}
}

func TestPrinter(t *testing.T) {
	p := NewPrinter()
	p.Print("a")
	p.Print("b", "hi", "yes")
	p.Indent()
	p.Print("cde")
	p.EndIndent()

	p.End()
	s := p.Fetch()
	expect := "a\nb hi yes {\n    cde\n}\n"
	if s != expect {
		t.Errorf("pson:\n %s\nexpect:\n %s", s, expect)
	}

	testToken(t, "a b", "'a b'")
	testToken(t, "a \n   b", "'a \n   b'")
	testToken(t, "a{", "'a{'")
	testToken(t, "}b", "'}b'")
	testToken(t, "'", "''''")
	testToken(t, "\\", "\\")
}
