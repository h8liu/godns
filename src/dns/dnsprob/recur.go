package dnsprob

import (
	. "dns"
	"fmt"
	"math/rand"
	"time"
)

// recursively query through the DNS hierarchy
type Recursive struct {
	n       *Name
	t       uint16
	start   *ZoneServers
	current *ZoneServers
	last    *ZoneServers
	Answer  *Msg
	AnsZone *ZoneServers
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

func shuffleServers(servers []*NameServer) []*NameServer {
	n := len(servers)
	ret := make([]*NameServer, n)
	order := rand.Perm(n)
	for i, ind := range order {
		ret[i] = servers[ind]
	}

	return ret
}

func shuffle(zs *ZoneServers) *ZoneServers {
	ret := []*NameServer{}
	nameOnly := []*NameServer{}

	for _, ns := range zs.Servers {
		if len(ns.Ips) == 0 {
			nameOnly = append(nameOnly, ns)
		} else {
			ret = append(ret, ns)
		}
	}

	ret = shuffleServers(ret)
	ret = append(ret, (shuffleServers(nameOnly))...)

	return &ZoneServers{zs.Zone, ret}
}

func NewRecursive(name *Name, t uint16) *Recursive {
	return &Recursive{
		n: name,
		t: t,
	}
}

func (p *Recursive) StartsWith(zone *ZoneServers) {
	p.start = zone
}

func (p *Recursive) Title() (name string, meta []string) {
	return "recur", []string{p.n.String(), TypeStr(p.t)}
}

func haveIP(ipList []*IPv4, ip *IPv4) bool {
	for _, i := range ipList {
		if i.Equal(ip) {
			return true
		}
	}
	return false
}

func (p *Recursive) nextZone(zs *ZoneServers) {
	if zs != nil {
		p.last = zs
	}
	p.current = zs
}

func (p *Recursive) queryZone(a ProbAgent) *Msg {
	zone := shuffle(p.current)
	tried := []*IPv4{}

	for _, server := range zone.Servers {
		ips := server.Ips
		if len(ips) == 0 {
			// ask for IPs here
			addr := NewAddr(server.Name)
			if !a.SolveSub(addr) {
				continue
			}
			ips = addr.Ips
			// nothing got
			if ips == nil {
				continue
			}

			if len(ips) == 0 {
				panic("ips got from Addr is empty set")
			}
		}

		for _, ip := range ips {
			if haveIP(tried, ip) {
				continue
			}
			tried = append(tried, ip)
			a.Log("//as",
				zone.Zone.String(),
				fmt.Sprintf("@%s(%s)", server.Name.String(), ip.String()))

			hisRecord := &QueryRecord{
				Host:   ip,
				Name:   p.n,
				Type:   p.t,
				Zone:   zone.Zone,
				Issued: time.Now(),
			}
			resp := a.Query(ip, p.n, p.t)
			hisRecord.Resp = resp
			p.History = append(p.History, hisRecord)

			if resp == nil {
				a.Log("//unreachable", server.Name.String())
				continue
			}

			msg := resp.Msg
			rcode := msg.Flags & F_RCODEMASK
			if !(rcode == RCODE_OKAY || rcode == RCODE_NAMEERROR) {
				a.Log("//svrerror", server.Name.String())
				continue
			}

			found, redirect := p.findAns(msg, a)
			if found {
				p.AnsCode = OKAY
				a.Log("//found")
				p.AnsZone = zone
				p.nextZone(nil)
				return msg // found
			} else {
				if redirect == nil {
					p.AnsCode = NONEXIST
					a.Log("//non-exist")
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

func (p *Recursive) findAns(msg *Msg, a ProbAgent) (bool, *ZoneServers) {
	// look for answer
	rrs := msg.FilterINRR(func(rr *RR, seg int) bool {
		if !rr.Name.Equal(p.n) {
			return false
		}
		return p.t == rr.Type || (p.t == A && rr.Type == CNAME)
	})
	if len(rrs) > 0 {
		return true, nil
	}

	// look for redirect name servers
	rrs = msg.FilterINRR(func(rr *RR, seg int) bool {
		if rr.Type != NS {
			return false
		}
		name := rr.Name
		if !name.SubOf(p.current.Zone) {
			return false
		}
		if !name.Equal(p.n) && !name.ParentOf(p.n) {
			return false
		}
		return true
	})
	if len(rrs) == 0 {
		// no record found, and no redirecting either
		return false, nil
	}

	subzone := rrs[0].Name // we only select the first subzone
	redirect := &ZoneServers{subzone, []*NameServer{}}

rrloop:
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
		for _, s := range redirect.Servers {
			if nsName.Equal(s.Name) {
				continue rrloop
			}
		}

		ns := &NameServer{nsName, []*IPv4{}}
		msg.FilterINRR(func(rr *RR, seg int) bool {
			if rr.Type != A || !rr.Name.Equal(nsName) {
				return false
			}
			ipData, ok := rr.Rdata.(*RdIP)
			if !ok || ipData.Ip == nil {
				return false
			}
			ns.Ips = append(ns.Ips, ipData.Ip)
			return false // handled already
		})

		redirect.Servers = append(redirect.Servers, ns)
	}

	if len(redirect.Servers) == 0 {
		panic("where are my redirect servers")
	}

	a.Cache(redirect)

	return false, redirect
}

var rootServers = makeRootServers()

func makeRootServers() *ZoneServers {
	ns := func(n string, ip string) *NameServer {
		return &NameServer{
			Name: MakeName(fmt.Sprintf("%s.root-servers.net", n)),
			Ips:  []*IPv4{ParseIP(ip)},
		}
	}

	// see en.wikipedia.org/wiki/Root_name_server for reference
	// (since year 2012)
	return &ZoneServers{Zone: MakeName("."),
		Servers: []*NameServer{
			ns("a", "192.41.0.4"),
			ns("b", "192.228.79.201"),
			ns("c", "192.33.4.12"),
			ns("d", "128.8.10.90"),
			ns("e", "192.203.230.10"),
			ns("f", "192.5.5.241"),
			ns("g", "192.112.36.4"),
			ns("h", "128.63.2.53"),
			ns("i", "192.36.148.17"),
			ns("j", "198.41.0.10"),
			ns("k", "193.0.14.129"),
			ns("l", "199.7.83.42"),
			ns("m", "202.12.27.33"),
		},
	}
}

func (p *Recursive) ExpandVia(a ProbAgent) {
	if p.start != nil {
		p.nextZone(p.start)
	} else {
		best := a.QueryCache(p.n)
		if best == nil {
			best = rootServers
		}
		p.nextZone(best)
	}

	p.History = make([]*QueryRecord, 0)
	for p.current != nil {
		p.Answer = p.queryZone(a)
	}
}
