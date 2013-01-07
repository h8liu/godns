package dns

import (
	"fmt"
	"time"
)

// recursively query through the DNS hierarchy
type ProbRecur struct {
	n       *Name
	t       uint16
	start   *Zone
	current *Zone
	last    *Zone
	Answer  *Msg
	AnsZone *Zone
	AnsCode int
	History []*QueryRecord
}

// to record the query history for recursive query problems
type QueryRecord struct {
	Host   *IPv4
	Name   *Name
	Type   uint16
	Zone   *Name
	Issued time.Time
	Resp   *Response
}

// answer codes
const (
	BUSY = 0
	OKAY = iota
	NONEXIST
	NORESP
)

func NewProbRecur(name *Name, t uint16) *ProbRecur {
	return &ProbRecur{
		n: name,
		t: t,
	}
}

func (p *ProbRecur) StartsWith(zone *Zone) {
	p.start = zone
}

func (p *ProbRecur) Title() (title []string) {
	return []string{"recur", p.n.String(), TypeStr(p.t)}
}

func haveIP(ipList []*IPv4, ip *IPv4) bool {
	for _, i := range ipList {
		if i.Equal(ip) {
			return true
		}
	}
	return false
}

func (p *ProbRecur) nextZone(zs *Zone) {
	if zs != nil {
		p.last = zs
	}
	p.current = zs
}

func (p *ProbRecur) queryZone(a Solver) *Msg {
	zone := p.current

	// prepare the servers 
	servers := zone.Prepare()
	tried := make(map[uint32]bool)

	for _, server := range servers {
		ips := server.IPs
		if len(ips) == 0 {
			// ask for IPs here
			addr := NewProbAddr(server.Name)
			if !a.SolveSub(addr) {
				continue
			}
			ips = addr.IPs
			// nothing got
			if ips == nil {
				continue
			}

			if len(ips) == 0 {
				panic("ips got from Addr is empty set")
			}
		}

		for _, ip := range ips {
			i := ip.Uint()
			if tried[i] {
				continue
			}
			tried[i] = true

			a.Log(fmt.Sprintf("// %s : %s(%s)",
				zone.Name().String(),
				server.Name.String(),
				ip.String(),
			))

			hisRecord := &QueryRecord{
				Host:   ip,
				Name:   p.n,
				Type:   p.t,
				Zone:   zone.Name(),
				Issued: time.Now(),
			}
			resp := a.Query(ip, p.n, p.t)
			hisRecord.Resp = resp
			p.History = append(p.History, hisRecord)

			if resp == nil {
				a.Log("// unreachable", server.Name.String())
				continue
			}

			msg := resp.Msg
			rcode := msg.Flags & F_RCODEMASK
			if !(rcode == RCODE_OKAY || rcode == RCODE_NAMEERROR) {
				a.Log("// server error", server.Name.String())
				continue
			}

			found, redirect := p.findAns(msg, a)
			if found {
				p.AnsCode = OKAY
				a.Log("// answer found")
				p.AnsZone = zone
				p.nextZone(nil)
				return msg // found
			} else {
				if redirect == nil {
					p.AnsCode = NONEXIST
					a.Log("// domain does not exist")
				}
				p.nextZone(redirect)
				return nil // found, but not exist
			}
		}
	}

	// got nothing, so set next zone to nil
	p.AnsCode = NORESP
	p.nextZone(nil)
	return nil
}

func (p *ProbRecur) findAns(msg *Msg, a Solver) (bool, *Zone) {
	// look for answer
	rrs := msg.FilterIN(func(rr *RR, seg int) bool {
		if !rr.Name.Equal(p.n) {
			return false
		}
		return p.t == rr.Type || (p.t == A && rr.Type == CNAME)
	})
	if len(rrs) > 0 {
		return true, nil
	}

	// look for redirect name servers
	rrs = msg.FilterIN(func(rr *RR, seg int) bool {
		if rr.Type != NS {
			return false
		}
		name := rr.Name
		if !name.SubOf(p.current.Name()) {
			return false // not under current zone, not trusted
		}
		return name.Equal(p.n) || name.ParentOf(p.n)
	})
	if len(rrs) == 0 {
		// no record found, and no redirecting either
		return false, nil
	}

	subzone := rrs[0].Name // we only select the first subzone
	redirect := NewZone(subzone)

	addedNSes := make(map[string]bool)
	addedIPs := make(map[uint32]bool)

	for _, rr := range rrs {
		if !rr.Name.Equal(subzone) {
			a.Log("warning:", "multiple subzones")
			continue
		}
		if rr.Class != IN || rr.Type != NS {
			panic("redirect record is wrong type")
		}
		nsData, ok := rr.Rdata.(*RdName)
		if !ok {
			panic("redirect record is not RdName")
		}

		nsName := nsData.Name
		nameStr := nsName.String()
		if addedNSes[nameStr] {
			continue
		}
		addedNSes[nameStr] = true

		redirect.AddName(nsName) // in case no IP is glued

		ips := make([]*IPv4, 0, 10)

		msg.ForEachIN(func(rr *RR, seg int) {
			if rr.Type != A || !rr.Name.Equal(nsName) {
				return
			}

			ipData, ok := rr.Rdata.(*RdIP)
			if !ok || ipData.IP == nil {
				return
			}

			i := ipData.IP.Uint()
			if addedIPs[i] {
				return
			}

			ips = append(ips, ipData.IP)
		})

		redirect.Add(nsName, ips...)

	}

	if IsRegistrar(subzone) {
		a.Log("// caching for zone:", subzone.String())
		a.Cache(redirect)
	}

	return false, redirect
}

var rootServers = makeRootServers()

func makeRootServers() *Zone {
	ret := NewZone(Domain("."))

	// helper function for adding servers
	ns := func(n string, ip string) {
		ret.Add(
			Domain(fmt.Sprintf("%s.root-servers.net", n)),
			ParseIP(ip),
		)
	}

	// see en.wikipedia.org/wiki/Root_name_server for reference
	// (since year 2012)
	// ns("a", "192.41.0.4") // Verisign
	ns("b", "192.228.79.201") // USC-ISI
	ns("c", "192.33.4.12")    // Cogent
	ns("d", "128.8.10.90")    // U Maryland
	ns("e", "192.203.230.10") // NASA
	ns("f", "192.5.5.241")    // Internet Systems Consortium
	ns("g", "192.112.36.4")   // DISA
	ns("h", "128.63.2.53")    // U.S. Army Research Lab
	ns("i", "192.36.148.17")  // Netnod
	ns("j", "198.41.0.10")    // Verisign
	ns("k", "193.0.14.129")   // RIPE NCC
	ns("l", "199.7.83.42")    // ICANN
	ns("m", "202.12.27.33")   // WIDE Project

	return ret
}

func (p *ProbRecur) ExpandVia(a Solver) {
	if p.start != nil {
		p.nextZone(p.start)
	} else {
		_, reg := RegParts(p.n)
		best := a.QueryCache(reg)
		if best == nil {
			best = rootServers
		}
		p.nextZone(best)
	}

	p.History = make([]*QueryRecord, 0, 50)
	for p.current != nil {
		p.Answer = p.queryZone(a)
	}
}
