package dns

import (
	"dns/pson"
	"fmt"
	"math/rand"
)

type Msg struct {
	ID    uint16
	Flags uint16
	Ques  []Ques
	Answ  []RR
	Auth  []RR
	Addi  []RR
}

type Ques struct {
	Name  *Name
	Type  uint16
	Class uint16
}

type RR struct {
	Name  *Name
	Type  uint16
	Class uint16
	TTL   uint32
	Rdata rdata
}

func NewQuesMsg(n *Name, t uint16) (ret *Msg) {
	ret = &Msg{0, 0,
		make([]Ques, 0),
		make([]RR, 0),
		make([]RR, 0),
		make([]RR, 0)}
	ret.Ques = append(ret.Ques, Ques{n, t, IN}) // copy in
	ret.RandID()

	return ret
}

func (m *Msg) FilterRR(f func (*RR, string) bool) []*RR {
    ret := []*RR{}
    for _, rr := range m.Answ {
        if f(&rr, ANSW) {
            ret = append(ret, &rr)
        }
    }
    for _, rr := range m.Auth {
        if f(&rr, AUTH) {
            ret = append(ret, &rr)
        }
    }
    for _, rr := range m.Addi {
        if f(&rr, ADDI) {
            ret = append(ret, &rr)
        }
    }
    return ret
}

func TypeStr(t uint16) string {
	switch t {
	case A:
		return "a"
	case NS:
		return "ns"
	case MD:
		return "md"
	case MF:
		return "mf"
	case CNAME:
		return "cname"
	case SOA:
		return "soa"
	case MB:
		return "mb"
	case MG:
		return "mg"
	case MR:
		return "mr"
	case NULL:
		return "null"
	case WKS:
		return "mks"
	case PTR:
		return "ptr"
	case HINFO:
		return "hinfo"
	case MINFO:
		return "minfo"
	case MX:
		return "mx"
	case TXT:
		return "txt"
	case AAAA:
		return "aaaa"
	}
	return fmt.Sprintf("t%d", t)
}

func ClassStr(t uint16) string {
	switch t {
	case IN:
		return "in"
	case CS:
		return "cs"
	case CH:
		return "ch"
	case HS:
		return "hs"
	}
	return fmt.Sprintf("c%d", t)
}

func TTLStr(t uint32) string {
	if t == 0 {
		return "0"
	}
	var ret string = ""
	sec := t % 60
	min := t / 60 % 60
	hour := t / 3600 % 24
	day := t / 3600 / 24
	if day > 0 {
		ret += fmt.Sprintf("%dd", day)
	}
	if hour > 0 {
		ret += fmt.Sprintf("%dh", hour)
	}
	if min > 0 {
		ret += fmt.Sprintf("%dm", min)
	}
	if sec > 0 {
		ret += fmt.Sprintf("%d", sec)
	}
	return ret
}

func (q *Ques) Pson(p *pson.Printer) {
	slist := make([]string, 0)
	if q.Type != A {
		slist = append(slist, TypeStr(q.Type))
	}
	if q.Class != IN {
		slist = append(slist, ClassStr(q.Type))
	}
	p.Print(q.Name.String(), slist...)
}

func (rr *RR) Pson(p *pson.Printer) {
	slist := make([]string, 0)

	slist = append(slist, TypeStr(rr.Type))
	rlist, expand := rr.Rdata.pson()
	for _, s := range rlist {
		slist = append(slist, s)
	}
	slist = append(slist, TTLStr(rr.TTL))
	if rr.Class != IN {
		slist = append(slist, ClassStr(rr.Class))
	}
	p.Print(rr.Name.String(), slist...)
	if expand {
		p.Indent()
		rr.Rdata.psonMore(p)
		p.EndIndent()
	}
}

func psonSection(p *pson.Printer, rrs []RR, sec string) {
	if len(rrs) == 0 {
		return
	}
	p.PrintIndent(sec)
	for _, rr := range rrs {
		rr.Pson(p)
	}
	p.EndIndent()
}

func (m *Msg) Pson(p *pson.Printer) {
	if (m.Flags & F_RESPONSE) == F_RESPONSE {
		p.PrintIndent("dns.resp")
	} else {
		p.PrintIndent("dns.query")
	}

	p.Print("id", fmt.Sprintf("%d", m.ID))
	fstr := make([]string, 0)
	switch {
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
		switch rcode {
		case RCODE_FORMATERROR:
			rs = "format-err"
		case RCODE_SERVERFAIL:
			rs = "server-fail"
		case RCODE_NAMEERROR:
			rs = "name-err"
		case RCODE_NOTIMPLEMENT:
			rs = "not-impl"
		case RCODE_REFUSED:
			rs = "refused"
		default:
			rs = fmt.Sprintf("unknown(%d)", rcode)
		}
		p.Print("rcode", rs)
	}

	if len(m.Ques) > 0 {
		p.PrintIndent("ques")
		for _, q := range m.Ques {
			q.Pson(p)
		}
		p.EndIndent()
	}

	psonSection(p, m.Answ, "answ")
	psonSection(p, m.Auth, "auth")
	psonSection(p, m.Addi, "addi")

	p.EndIndent()
}

func (m *Msg) String() string {
	p := pson.NewPrinter()
	m.Pson(p)
	p.End()

	return p.Fetch()
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
