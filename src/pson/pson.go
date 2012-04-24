package pson

import (
	"io"
)

type Encoder struct {
	out io.Writer
	indent uint
	ntoken uint
}

func NewEncoder (out io.Writer) *Encoder{
	return &Encoder{out, 0, 0}
}

func (e *Encoder) Add(s string, args ...string) (err error){
	return e.AddWith(s, args)
}

func (e *Encoder) emit(s string) (err error) {
	_, err = e.out.Write([]byte(s)); if err != nil { return err }
	return nil
}

func token(t string) (s string, err error) {
	// TODO
	return t, nil
}

func (e *Encoder) emitToken(t string) (err error) {
	t, err = token(t); if err != nil { return err }
	err = e.emit(t); if err != nil { return err }
	return nil
}

func (e *Encoder) EndLine() (err error) {
	_, err = e.out.Write([]byte("\n")); if err != nil { return err }
	e.ntoken = 0
	return nil
}

func (e *Encoder) AddWith(s string, args []string) (err error) {
	if e.ntoken != 0 {
		err = e.EndLine(); if err != nil { return err }
	}

	err = e.emitToken(s); if err != nil { return err }
	for _, a := range args {
		err = e.emitToken(a); if err != nil { return err }
	}
	return nil
}

func (e *Encoder) Sub() (err error) {
	err = e.emit("{"); if err != nil { return err }
	err = e.EndLine(); if err != nil { return err }
	e.indent++
	return nil
}

func (e *Encoder) EndSub() (err error) {
	if e.indent == 0 {
		return nil // no effect
	}

	if e.ntoken != 0 {
		err = e.EndLine(); if err != nil { return err }
	}
	e.indent--
	e.emit("}")
	err = e.EndLine(); if err != nil { return err }
	return nil
}

