package pson

import (
	"io"
	"strings"
)

type Printer struct {
	out    io.Writer
	indent uint
	ntoken uint
}

func NewPrinter(out io.Writer) *Printer {
	return &Printer{out, 0, 0}
}

func (e *Printer) emit(s string) (err error) {
	t := Tokenize(s)
	err = e.emitToken(t)
	if err != nil {
		return err
	}
	return nil
}

func isNormal(s string) bool {
	for _, c := range s {
		if !(c >= '!' && c <= '~') {
			return false
		}
		switch {
		case c == '\'':
			return false
		case c == '\\':
			return false
		case c == '}' || c == '{':
			return false
		}
	}
	return true
}

func Tokenize(s string) (t string) {
	if isNormal(s) {
		return s
	}
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "'", "\\'", -1)
	return "'" + s + "'"
}

func (e *Printer) emitToken(t string) (err error) {
	if e.ntoken == 0 {
		for i := uint(0); i < e.indent; i++ {
			e.out.Write([]byte("    "))
		}
	} else {
		e.out.Write([]byte(" "))
	}
	_, err = e.out.Write([]byte(t))
	if err != nil {
		return err
	}
	e.ntoken++
	return nil
}

func (e *Printer) EndLine() (err error) {
	_, err = e.out.Write([]byte("\n"))
	if err != nil {
		return err
	}
	e.ntoken = 0
	return nil
}

func (e *Printer) Print(s string, args ...string) (err error) {
	if e.ntoken != 0 {
		err = e.EndLine()
		if err != nil {
			return err
		}
	}

	err = e.emit(s)
	if err != nil {
		return err
	}
	for _, a := range args {
		err = e.emit(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Printer) PrintIndent(s string, args ...string) (err error) {
	err = e.Print(s, args...)
	if err != nil {
		return err
	}
	err = e.Indent()
	if err != nil {
		return err
	}
	return nil
}

func (e *Printer) Indent() (err error) {
	err = e.emitToken("{")
	if err != nil {
		return err
	}
	err = e.EndLine()
	if err != nil {
		return err
	}
	e.indent++
	return nil
}

func (e *Printer) EndIndent() (err error) {
	if e.indent == 0 {
		return nil // no effect
	}

	if e.ntoken != 0 {
		err = e.EndLine()
		if err != nil {
			return err
		}
	}
	e.indent--
	e.emitToken("}")
	err = e.EndLine()
	if err != nil {
		return err
	}
	return nil
}

func (e *Printer) End() (err error) {
	if e.ntoken != 0 {
		return e.EndLine()
	}
	return nil
}
