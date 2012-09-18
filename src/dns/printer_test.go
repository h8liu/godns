package dns

import (
	"testing"
)

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
}
