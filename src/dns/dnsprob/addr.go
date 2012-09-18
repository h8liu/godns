package dnsprob

import . "dns"

type Addr struct {
	name *Name
	Ips  []*IPv4
}

func NewAddr(name *Name) *Addr {
	return &Addr{name, nil}
}

func (p *Addr) Title() (name string, meta []string) {
	return "addr", []string{p.name.String()}
}

func (p *Addr) ExpandVia(a Agent) {
	recur := NewRecursive(p.name, A)
	if !a.SolveSub(recur) {
		return
	}

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
