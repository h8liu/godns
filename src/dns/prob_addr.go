package dns

type ProbAddr struct {
	name *Name
	IPs  []*IPv4
}

func NewProbAddr(name *Name) *ProbAddr {
	return &ProbAddr{name, nil}
}

func (p *ProbAddr) Title() (title []string) {
	return []string{"addr", p.name.String()}
}

func (p *ProbAddr) ExpandVia(a Solver) {
	recur := NewProbRecur(p.name, A)
	if !a.SolveSub(recur) {
		return
	}

	ans := recur.Answer
	if ans == nil {
		return
	}

	/* first find A records */
	rrs := ans.FilterIN(func(rr *RR, seg int) bool {
		return rr.Name.Equal(p.name) && rr.Type == A
	})

	if len(rrs) > 0 {
		p.IPs = toIPs(rrs)
		return
	}

	// not found, then look for cnames
	rrs = ans.FilterIN(func(rr *RR, seg int) bool {
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
	rrs = ans.FilterIN(func(rr *RR, seg int) bool {
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
	p.IPs = toIPs(rrs)
}

func toIPs(rrs []*RR) []*IPv4 {
	ret := make([]*IPv4, len(rrs))
	for i, rr := range rrs {
		ret[i] = rr.Rdata.(*RdIP).IP
	}
	return ret
}
