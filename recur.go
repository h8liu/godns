package dns

import (
	"fmt"
	"math/rand"
)

// recursively query through the DNS hierarchy
type RecurProb struct {
	n           *Name
	t           uint16
	start       *ZoneServers
	current     *ZoneServers
	last        *ZoneServers
	Answer      *Msg
	ansZone     *ZoneServers
	ansServerIP *IPv4
}

type ZoneServers struct {
	Zone    *Name
	Servers []*NameServer
}

type NameServer struct {
	Name *Name
	Ips  []*IPv4
}

func shuffleServers(servers []*NameServer) []*NameServer {
	n := len(servers)
	ret := make([]*NameServer, n)
	order := rand.Perm(n)
	for i, ind := range order {
		ret[i] = servers[ind]
	}

	return ret
}

func (zs *ZoneServers) shuffle() *ZoneServers {
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
	nameOnly = shuffleServers(nameOnly)
	for _, ns := range nameOnly {
		ret = append(ret, ns)
	}

	return &ZoneServers{zs.Zone, ret}
}

func NewRecurProb(name *Name, t uint16) *RecurProb {
	ret := new(RecurProb)
	ret.n = name
	ret.t = t
	ret.Answer = nil

	return ret
}

func (p *RecurProb) StartFrom(zone *Name, servers []*NameServer) {
	p.start = &ZoneServers{zone, servers}
}

func (p *RecurProb) Title() (name string, meta []string) {
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

func (p *RecurProb) nextZone(zs *ZoneServers) {
	if zs != nil {
		p.last = zs
	}
	p.current = zs
}

func (p *RecurProb) queryZone(a Agent) *Msg {
	zone := p.current.shuffle()
	tried := []*IPv4{}

	for _, server := range zone.Servers {
		if len(server.Ips) == 0 {
			// TODO: ask IP first. will do this after AddrProb is done
		}
		if len(server.Ips) == 0 {
			continue
		}
		for _, ip := range server.Ips {
			if haveIP(tried, ip) {
				continue
			}
			tried = append(tried, ip)
			a.Log("//as",
				zone.Zone.String(),
				fmt.Sprintf("@%s(%s)", server.Name.String(), ip.String()))
			resp := a.Query(ip, p.n, p.t)
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
				a.Log("//found")
				p.ansZone = &ZoneServers{
					zone.Zone,
					[]*NameServer{&NameServer{
						server.Name,
						[]*IPv4{ip},
					}},
				}
				p.nextZone(nil)
				return msg
			} else {
				if redirect == nil {
					a.Log("//non-exist")
				}
				p.nextZone(redirect)
				return nil
			}
		}
	}

	// got nothing, so set next zone to nil
	p.nextZone(nil)
	return nil
}

func (p *RecurProb) findAns(msg *Msg, a Agent) (bool, *ZoneServers) {
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
			Name: makeName(fmt.Sprintf("%s.root-servers.net", n)),
			Ips:  []*IPv4{ParseIP(ip)},
		}
	}

	// see en.wikipedia.org/wiki/Root_name_server for reference
	// (since year 2012)
	return &ZoneServers{Zone: makeName("."),
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

func (p *RecurProb) ExpandVia(a Agent) {
	if p.start != nil {
		p.nextZone(p.start)
	} else {
		best := a.QueryCache(p.n)
		if best == nil {
			best = rootServers
		}
		p.nextZone(best)
	}

	for p.current != nil {
		p.Answer = p.queryZone(a)
	}
}

func (p *RecurProb) Prob() Prob {
	return p
}
