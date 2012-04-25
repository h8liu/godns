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

type RdBytes struct {
	data []byte
}

func (rd *RdBytes) pson() ([]string, bool) {
	return []string{}, false
}

func (rd *RdBytes) psonMore(p *pson.StrPrinter) {
}

func (rd *RdBytes) writeTo(w *writer) error {
	w.writeUint16(0)
	return nil
}

func (rd *RdBytes) readFrom(r *reader, n uint16) error {
	rd.data = make([]byte, n)
	return r.readBytes(rd.data)
}

type RdIP struct {
	ip IPv4
}

// should be treated as immutable
type IPv4 [4]byte

func (ip *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip[0], ip[1], ip[2], ip[3])
}

func (rd *RdIP) pson() ([]string, bool) {
	return []string{rd.ip.String()}, false
}

func (rd *RdIP) psonMore(p *pson.StrPrinter) {
}

func (rd *RdIP) writeTo(w *writer) error {
	w.writeBytes(rd.ip[:])
	return nil
}

func (rd *RdIP) readFrom(r *reader, n uint16) (err error) {
	if n != 4 {
		return &ParseError{"A rdata: wrong size"}
	}
	err = r.readBytes(rd.ip[:])
	if err != nil {
		return err
	}
	return nil
}

type RdName struct {
	name *Name
}

func (r *RdName) pson() ([]string, bool) {
	return []string{r.name.String()}, false
}

func (rd *RdName) psonMore(p *pson.StrPrinter) {
}

func (rd *RdName) writeTo(w *writer) error {
	panic("not implemented")
	w.writeName(rd.name)
	return nil
}

func (rd *RdName) readFrom(r *reader, n uint16) (err error) {
	rd.name, err = r.readName()
	if err != nil {
		return err
	}
	return nil
}