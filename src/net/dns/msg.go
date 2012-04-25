package dns

import (
	"fmt"
	"math/rand"
	"pson"
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
	case t == AAAA:
		return "aaaa"
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

func (q *Ques) Pson(p *pson.StrPrinter) {
	if q.Class == IN {
		p.Print(q.Name.String(), TypeString(q.Type))
	} else {
		p.Print(q.Name.String(), TypeString(q.Type),
			ClassString(q.Class))
	}
}

func (rr *RR) Pson(p *pson.StrPrinter) {
	// TODO: print RData
	if rr.Class == IN {
		p.Print(rr.Name.String(), TypeString(rr.Type),
			fmt.Sprintf("%d", rr.TTL))
	} else {
		p.Print(rr.Name.String(), TypeString(rr.Type),
			fmt.Sprintf("%d", rr.TTL), ClassString(rr.Class))
	}
}

func psonSection(p *pson.StrPrinter, rrs []RR, sec string) {
	if len(rrs) == 0 {
		return
	}
	p.PrintIndent(sec)
	for _, rr := range rrs {
		rr.Pson(p)
	}
	p.EndIndent()
}

func (m *Msg) Pson(p *pson.StrPrinter) {
	p.Print("id", fmt.Sprintf("%d", m.ID))
	fstr := make([]string, 0)
	switch {
	case (m.Flags & F_RESPONSE) == F_RESPONSE:
		fstr = append(fstr, "resp")
	case (m.Flags & F_OPMASK) == OPIQUERY:
		fstr = append(fstr, "op=iquery")
	case (m.Flags & F_OPMASK) == OPSTATUS:
		fstr = append(fstr, "op=status")
	case (m.Flags & F_AA) == F_AA:
		fstr = append(fstr, "auth")
	case (m.Flags & F_TC) == F_TC:
		fstr = append(fstr, "trunc")
	case (m.Flags & F_RD) == F_RD:
		fstr = append(fstr, "rec-desired")
	case (m.Flags & F_RA) == F_RA:
		fstr = append(fstr, "rec-avail")
	}
	if len(fstr) > 0 {
		p.Print("flag", fstr...)
	}
	rcode := m.Flags & F_RCODEMASK
	if rcode != RCODE_OKAY {
		var rs string
		switch {
		case rcode == RCODE_FORMATERROR:
			rs = "format-err"
		case rcode == RCODE_SERVERFAIL:
			rs = "server-fail"
		case rcode == RCODE_NAMEERROR:
			rs = "name-err"
		case rcode == RCODE_NOTIMPLEMENT:
			rs = "not-impl"
		case rcode == RCODE_REFUSED:
			rs = "refused"
		default:
			rs = fmt.Sprintf("unknown(%d)", rcode)
		}
		p.Print("rcode", rs)
	}

	if len(m.Ques) > 0 {
		p.PrintIndent("question")
		for _, q := range m.Ques {
			q.Pson(p)
		}
		p.EndIndent()
	}

	psonSection(p, m.Answ, "answer")
	psonSection(p, m.Auth, "authority")
	psonSection(p, m.Addi, "additional")
}

func (m *Msg) String() string {
	p := pson.NewStrPrinter()
	p.PrintIndent("dns.msg")
	m.Pson(p)
	p.EndIndent()
	return p.End()
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