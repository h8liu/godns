package pson

import (
	"bytes"
)

type StrPrinter struct {
	buf bytes.Buffer
	p   *Printer
}

func NewStrPrinter() *StrPrinter {
	ret := new(StrPrinter)
	ret.p = NewPrinter(&ret.buf)
	return ret
}

func (e *StrPrinter) EndLine() {
	err := e.p.EndLine()
	if err != nil {
		panic(err)
	}
}

func (e *StrPrinter) Print(s string, args ...string) {
	err := e.p.Print(s, args...)
	if err != nil {
		panic(err)
	}
}

func (e *StrPrinter) Indent() {
	err := e.p.Indent()
	if err != nil {
		panic(err)
	}
}
func (e *StrPrinter) PrintIndent(s string, args ...string) {
	err := e.p.PrintIndent(s, args...)
	if err != nil {
		panic(err)
	}
}

func (e *StrPrinter) EndIndent() {
	err := e.p.EndIndent()
	if err != nil {
		panic(err)
	}
}

func (e *StrPrinter) End() string {
	err := e.p.End()
	if err != nil {
		panic(err)
	}
	return string(e.buf.Bytes())
}
