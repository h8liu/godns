package dns

type AddrProb struct {
	name *Name
	Ips  []*IPv4
}

func NewAddrProb(name *Name) *AddrProb {
	return &AddrProb{name, nil}
}

func (p *AddrProb) Title() (name string, meta []string) {
	return "addr", []string{p.name.String()}
}

func (p *AddrProb) ExpandVia(a Agent) {
	recur := NewRecurProb(p.name, A)
	a.SolveSub(recur)
	ans := recur.Answer
	if ans == nil {
		return
	}

	/* first find A records */
	rrs := ans.FilterINRR(func(rr *RR, seg int) bool {
		return rr.Name.Equal(p.name) && rr.Type == A
	})

	if len(rrs) > 0 {
		p.Ips = toIps(rrs)
		return
	}

	// not found, then look for cnames
	rrs = ans.FilterINRR(func(rr *RR, seg int) bool {
		return rr.Name.Equal(p.name) && rr.Type == CNAME
	})
	if len(rrs) == 0 {
		return
	}

	cnames := make([]*Name, len(rrs))
	for i, rr := range rrs {
		cnames[i] = rr.Rdata.(*RdName).Name
	}

	// look for glued ips
	rrs = ans.FilterINRR(func(rr *RR, seg int) bool {
		if rr.Type != A {
			return false
		}
		name := rr.Name
		for _, n := range cnames {
			n.Equal(name)
			return true
		}
		return false
	})

	if len(rrs) == 0 {
		return
	}
	p.Ips = toIps(rrs)
}

func toIps(rrs []*RR) []*IPv4 {
	ret := make([]*IPv4, len(rrs))
	for i, rr := range rrs {
		ret[i] = rr.Rdata.(*RdIP).Ip
	}
	return ret
}
