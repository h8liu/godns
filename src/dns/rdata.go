package dns

import (
	"fmt"
	"pson"
    "net"
)

type rdata interface {
	pson() ([]string, bool)
	psonMore(p *pson.Printer)
	writeTo(w *writer) error
	readFrom(r *reader, n uint16) error
}

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

type RdIP struct {
	ip IPv4
}

// should be treated as immutable
type IPv4 [4]byte

func (ip *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip[0], ip[1], ip[2], ip[3])
}

func ParseIP(s string) *IPv4 {
    nip := net.ParseIP(s)
    if nip == nil { 
        return nil
    }
    nip = nip.To4()
    if nip == nil {
        return nil
    }
    
    ret := new(IPv4)
    copy(ret[:], nip)
    return ret
}

func (ip *IPv4) Equal(other *IPv4) bool {
	for i, b := range ip {
		if b != other[i] {
			return false
		}
	}
	return true
}

func (rd *RdIP) pson() ([]string, bool) {
	return []string{rd.ip.String()}, false
}

func (rd *RdIP) psonMore(p *pson.Printer) {
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

func (rr *RR) RdIP() *RdIP {
	if rr.Class == IN {
		if rr.Type == A {
			return rr.rdata.(*RdIP)
		}
	}
	return nil
}

func (rr *RR) RdName() *RdName {
	if rr.Class == IN {
		switch rr.Type {
		case CNAME, NS:
			return rr.rdata.(*RdName)
		}
	}
	return nil
}

func (rr *RR) RdBytes() *RdBytes {
	if rr.Class == IN {
		switch rr.Type {
		case TXT:
			return rr.rdata.(*RdBytes)
		}
	}
	return nil
}
