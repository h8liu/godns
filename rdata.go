package dns

import (
	"dns/pson"
)

type rdata interface {
	pson() ([]string, bool)
	psonMore(p *pson.Printer)
	writeTo(w *writer) error
	readFrom(r *reader, n uint16) error
}

// for rdata of a string of a byte array, like txt records
type RdBytes struct {
	data []byte
}

func (rd *RdBytes) pson() ([]string, bool) {
	return []string{}, false
}

func (rd *RdBytes) psonMore(p *pson.Printer) {
}

func (rd *RdBytes) writeTo(w *writer) error {
	w.writeUint16(0)
	return nil
}

func (rd *RdBytes) readFrom(r *reader, n uint16) error {
	rd.data = make([]byte, n)
	return r.readBytes(rd.data)
}

// for rdatas of a single ip address, like a records
type RdIP struct {
	ip *IPv4
}

func (rd *RdIP) pson() ([]string, bool) {
	return []string{rd.ip.String()}, false
}

func (rd *RdIP) psonMore(p *pson.Printer) {
}

func (rd *RdIP) writeTo(w *writer) error {
	w.writeBytes(rd.ip.Bytes())
	return nil
}

func (rd *RdIP) readFrom(r *reader, n uint16) (err error) {
	if n != 4 {
		return &ParseError{"A rdata: wrong size"}
	}
	buf := make([]byte, 4)
	if err = r.readBytes(buf); err != nil {
		return err
	}
	if rd.ip = IPFromBytes(buf); rd.ip == nil {
		return &ParseError{"make ip from bytes"}
	}
	return nil
}

// for rdatas of a single name, like ns records
type RdName struct {
	name *Name
}

func (r *RdName) pson() ([]string, bool) {
	return []string{r.name.String()}, false
}

func (rd *RdName) psonMore(p *pson.Printer) {
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
