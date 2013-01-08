package dns

import (
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
	Rdata Rdata
}

func NewQuery(n *Name, t uint16) (ret *Msg) {
	ret = &Msg{0, 0,
		make([]Ques, 0, 1),
		make([]RR, 0, 10),
		make([]RR, 0, 10),
		make([]RR, 0, 10)}
	ret.Ques = append(ret.Ques, Ques{n, t, IN}) // copy in
	ret.RollAnID()

	return ret
}

func (m *Msg) Filter(f func(*RR, int) bool) []*RR {
	ret := make([]*RR, 0, 30)
	for i := 0; i < len(m.Answ); i++ {
		rr := &m.Answ[i]
		if f(rr, ANSW) {
			ret = append(ret, rr)
		}
	}

	for i := 0; i < len(m.Auth); i++ {
		rr := &m.Auth[i]
		if f(rr, AUTH) {
			ret = append(ret, rr)
		}
	}

	for i := 0; i < len(m.Addi); i++ {
		rr := &m.Addi[i]
		if f(rr, ADDI) {
			ret = append(ret, rr)
		}
	}

	return ret
}

func (m *Msg) FilterIN(f func(*RR, int) bool) []*RR {
	return m.Filter(func(rr *RR, seg int) bool {
		if rr.Class != IN {
			return false
		}
		return f(rr, seg)
	})
}

func (m *Msg) ForEach(f func(*RR, int)) {
	m.Filter(func(rr *RR, seg int) bool {
		f(rr, seg)
		return false
	})
}

func (m *Msg) ForEachIN(f func(*RR, int)) {
	m.FilterIN(func(rr *RR, seg int) bool {
		f(rr, seg)
		return false
	})
}

var typeStrs = map[uint16]string{
	A:     "a",
	NS:    "ns",
	MD:    "md",
	MF:    "mf",
	CNAME: "cname",
	SOA:   "soa",
	MB:    "mb",
	MG:    "mg",
	MR:    "mr",
	NULL:  "null",
	HINFO: "hinfo",
	MINFO: "minfo",
	MX:    "mx",
	TXT:   "txt",
	AAAA:  "aaaa",
}

func TypeStr(t uint16) string {
	ret, has := typeStrs[t]
	if has {
		return ret
	}
	return fmt.Sprintf("t%d", t)
}

var classStrs = map[uint16]string{
	IN: "in",
	CS: "cs",
	CH: "ch",
	HS: "hs",
}

func ClassStr(t uint16) string {
	ret, has := classStrs[t]
	if has {
		return ret
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

func (q *Ques) printTo(p *printer) {
	slist := make([]string, 0, 5)
	slist = append(slist, q.Name.String())
	if q.Type != A {
		slist = append(slist, TypeStr(q.Type))
	}
	if q.Class != IN {
		slist = append(slist, ClassStr(q.Type))
	}
	p.Print(slist...)
}

func (rr *RR) printTo(p *printer) {
	slist := make([]string, 0, 10)
	slist = append(slist, rr.Name.String())
	slist = append(slist, TypeStr(rr.Type))
	rlist, expand := rr.Rdata.printOut()
	slist = append(slist, rlist...)
	slist = append(slist, TTLStr(rr.TTL))
	if rr.Class != IN {
		slist = append(slist, ClassStr(rr.Class))
	}
	p.Print(slist...)
	if expand != nil {
		p.Indent()
		expand(p)
		p.EndIndent()
	}
}

func printSection(p *printer, rrs []RR, sec string) {
	if len(rrs) == 0 {
		return
	}
	p.PrintIndent(sec)
	for _, rr := range rrs {
		rr.printTo(p)
	}
	p.EndIndent()
}

func (m *Msg) printTo(p *printer) {
	if (m.Flags & F_RESPONSE) != F_RESPONSE {
		p.Print("//query")
	}

	fstr := make([]string, 0, 5)
	fstr = append(fstr, fmt.Sprintf("#%d", m.ID))
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
		p.Print(fstr...)
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
			q.printTo(p)
		}
		p.EndIndent()
	}

	printSection(p, m.Answ, "answ")
	printSection(p, m.Auth, "auth")
	printSection(p, m.Addi, "addi")
}

func (m *Msg) String() string {
	p := newPrinter()
	m.printTo(p)
	p.End()

	return p.Fetch()
}

func (m *Msg) RollAnID() {
	m.ID = uint16(rand.Uint32())
}

func (m *Msg) Wire() ([]byte, error) {
	w := new(writer)
	e := w.writeMsg(m)
	if e != nil {
		return nil, e
	}
	return w.wire(), nil
}

func ParseMsg(buf []byte) (*Msg, error) {
	r := newReader(buf)
	ret := new(Msg)
	e := r.readMsg(ret)
	if e != nil {
		return nil, e
	}
	return ret, nil
}

func (rr *RR) String() string {
	p := newPrinter()
	rr.printTo(p)
	p.End()
	return p.Fetch()
}
