package dns

import (
	"bytes"
	"io"
)

type Printer struct {
	out    bytes.Buffer
	indent uint
	ntoken uint
}

func NewPrinter() *Printer {
	return new(Printer)
}

func (e *Printer) FlushTo(out io.Writer) (n int, err error) {
	n, err = out.Write(e.out.Bytes())
	if err != nil {
		return
	}
	e.out.Reset()
	return
}

func (e *Printer) Fetch() string {
	ret := string(e.out.Bytes())
	e.out.Reset()
	return ret
}

func (e *Printer) emitString(s string) {
	e.emitToken(s)
}

func (e *Printer) emitToken(t string) {
	if e.ntoken == 0 {
		for i := uint(0); i < e.indent; i++ {
			e.out.Write([]byte("    "))
		}
	} else {
		e.out.Write([]byte(" "))
	}
	e.out.Write([]byte(t))
	e.ntoken++
}

func (e *Printer) EndLine() {
	e.out.Write([]byte("\n"))
	e.ntoken = 0
}

func (e *Printer) Print(args ...string) {
	if e.ntoken != 0 {
		e.EndLine()
	}
	for _, a := range args {
		e.emitString(a)
	}
}

func (e *Printer) PrintIndent(args ...string) {
	e.Print(args...)
	e.Indent()
}

func (e *Printer) Indent() {
    e.emitToken("{")
	e.EndLine()
	e.indent++
}

func (e *Printer) EndIndent() {
	if e.indent == 0 {
		return // no effect
	}
	if e.ntoken != 0 {
		e.EndLine()
	}
	e.indent--
    e.emitToken("}")
    e.EndLine()
}

func (e *Printer) End() {
	if e.ntoken != 0 {
		e.EndLine()
	}

	for e.indent > 0 {
		e.EndIndent()
	}
}
