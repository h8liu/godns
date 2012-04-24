package dns

import (
	"math/rand"
	"pson"
	"bytes"
	"fmt"
)

type Msg struct {
	ID    uint16
	Flags uint16
	Ques  []Ques
	Answ  []RR
	Auth  []RR
	Addi  []RR
}

func QuesMsg(n string, t uint16) (ret *Msg, err error) {
	name, e := NewName(n)
	if e != nil {
		return nil, e
	}
	ret = &Msg{0, 0,
		make([]Ques, 0),
		make([]RR, 0),
		make([]RR, 0),
		make([]RR, 0)}
	ret.Ques = append(ret.Ques, Ques{name, t, IN}) // copy in
	ret.RandID()

	return ret, nil
}

func TypeString(t uint16) string {
	switch {
	case t == A:
		return "a"
	case t == NS:
		return "ns"
	case t == MD:
		return "md"
	case t == MF:
		return "mf"
	case t == CNAME:
		return "cname"
	case t == SOA:
		return "soa"
	case t == MB:
		return "mb"
	case t == MG:
		return "mg"
	case t == MR:
		return "mr"
	case t == NULL:
		return "null"
	case t == WKS:
		return "mks"
	case t == PTR:
		return "ptr"
	case t == HINFO:
		return "hinfo"
	case t == MINFO:
		return "minfo"
	case t == MX:
		return "mx"
	case t == TXT:
		return "txt"
	}
	return "-"
}

func ClassString(t uint16) string {
	switch {
	case t == IN:
		return "in"
	case t == CS:
		return "cs"
	case t == CH:
		return "ch"
	case t == HS:
		return "hs"
	}
	return "-"
}

func (q *Ques) Pson(p *pson.Printer) error {
	if q.Class == IN {
		return p.Print(q.Name.String(), TypeString(q.Type))
	}
	return p.Print(q.Name.String(), TypeString(q.Type),
		ClassString(q.Class))
}

func (rr *RR) Pson(p *pson.Printer) error {
	if rr.Class == IN {
		return p.Print(rr.Name.String(), TypeString(rr.Type),
			fmt.Sprintf("%d", rr.TTL))
	}

	return p.Print(rr.Name.String(), TypeString(rr.Type),
		fmt.Sprintf("%d", rr.TTL), ClassString(rr.Class))
}

func psonSection(p *pson.Printer, rrs []RR, sec string) (e error) {
	fmt.Printf("%s %d\n", sec, len(rrs))
	if len(rrs) == 0 { return nil }
	e = p.PrintIndent(sec); if e != nil { return }
	for _, rr := range rrs {
		e = rr.Pson(p); if e != nil { return }
	}
	e = p.EndIndent(); if e != nil { return }
	return nil
}


func (m *Msg) Pson(p *pson.Printer) (e error){
	e = p.Print("id", fmt.Sprintf("%d", m.ID)); if e != nil { return }

	if len(m.Ques) > 0 {
		e = p.PrintIndent("question"); if e != nil { return }
		for _, q := range m.Ques {
			e = q.Pson(p); if e != nil { return }
		}
		e = p.EndIndent(); if e != nil { return }
	}

	e = psonSection(p, m.Answ, "answer"); if e != nil { return }
	e = psonSection(p, m.Auth, "authority"); if e != nil { return }
	e = psonSection(p, m.Addi, "additional"); if e != nil { return }

	return nil
}

func (m *Msg) String() string {
	buf := new(bytes.Buffer)
	p := pson.NewPrinter(buf)
	p.Print("dns.msg")
	p.Indent()

	m.Pson(p)

	p.EndIndent()
	p.End()

	return string(buf.Bytes())
}

func (m *Msg) RandID() {
	m.ID = uint16(rand.Uint32())
}

func (m *Msg) ToWire() ([]byte, error) {
	w := new(writer)
	e := w.writeMsg(m)
	if e != nil {
		return nil, e
	}
	return w.wire(), nil
}

func FromWire(buf []byte) (*Msg, error) {
	r := newReader(buf)
	ret := new(Msg)
	e := r.readMsg(ret)
	if e != nil {
		return nil, e
	}
	return ret, nil
}
