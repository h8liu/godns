package dns

import (
	"fmt"
	"pson"
)

type rdata interface {
	pson() ([]string, bool)
	psonMore(p *pson.StrPrinter)
	writeTo(w *writer) error
	readFrom(r *reader, n uint16) error
}

type RdAny struct {
	data []byte
}

func (rd *RdAny) pson() ([]string, bool) {
	return []string{}, false
}

func (rd *RdAny) psonMore(p *pson.StrPrinter) {
}

func (rd *RdAny) writeTo(w *writer) error {
	w.writeUint16(0)
	return nil
}

func (rd *RdAny) readFrom(r *reader, n uint16) error {
	rd.data = make([]byte, n)
	return r.readBytes(rd.data)
}

type RdA struct {
	ip IPv4
}

// should be treated as immutable
type IPv4 [4]byte

func (ip *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip[0], ip[1], ip[2], ip[3])
}

func (rd *RdA) pson() ([]string, bool) {
	return []string{rd.ip.String()}, false
}

func (rd *RdA) psonMore(p *pson.StrPrinter) {
}

func (rd *RdA) writeTo(w *writer) error {
	w.writeBytes(rd.ip[:])
	return nil
}

func (rd *RdA) readFrom(r *reader, n uint16) (err error) {
	if n != 4 {
		return &ParseError{"A rdata: wrong size"}
	}
	err = r.readBytes(rd.ip[:])
	if err != nil {
		return err
	}
	return nil
}

type RdNS struct {
	name *Name
}

func (r *RdNS) pson() ([]string, bool) {
	return []string{r.name.String()}, false
}

func (rd *RdNS) psonMore(p *pson.StrPrinter) {
}

func (rd *RdNS) writeTo(w *writer) error {
	panic("not implemented")
	w.writeName(rd.name)
	return nil
}

func (rd *RdNS) readFrom(r *reader, n uint16) (err error) {
	rd.name, err = r.readName()
	if err != nil {
		return err
	}
	return nil
}
