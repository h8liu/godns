package dns

import (
	"fmt"
	"math/rand"
)

// recursively query a question through a bunch of servers
// only focus on one single record type
type RecurProb struct {
	n       *Name
	t       uint16
	nscache *NSCache
	start   *ZoneServers
	current *ZoneServers
	last    *ZoneServers
	answer  *Msg
}

type ZoneServers struct {
	zone    *Name
	servers []*NameServer
}

type NameServer struct {
	name *Name
	ips  []*IPv4
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

func (zs *ZoneServers) shuffle() {
	res := []*NameServer{}
	nameOnly := []*NameServer{}

	for _, ns := range zs.servers {
		if len(ns.ips) == 0 {
			nameOnly = append(nameOnly, ns)
		} else {
			res = append(res, ns)
		}
	}

	res = shuffleServers(res)
	nameOnly = shuffleServers(nameOnly)
	for _, ns := range nameOnly {
		res = append(res, ns)
	}
	zs.servers = res
}

func NewRecurProb(name *Name, t uint16) *RecurProb {
	ret := new(RecurProb)
	ret.n = name
	ret.t = t
	ret.nscache = nil
	ret.answer = nil

	ret.UseCache(globalNSCache)
	ret.StartFromRoot()

	return ret
}

func (p *RecurProb) StartFromRoot() {
	p.start = nil
}

func (p *RecurProb) StartFrom(zone *Name, servers []*NameServer) {
	p.start = &ZoneServers{zone, servers}
}

func (p *RecurProb) UseCache(cache *NSCache) {
	p.nscache = cache
}

func (p *RecurProb) Title() (name string, meta []string) {
	return "rec", []string{p.n.String(), TypeStr(p.t)}
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

func (p *RecurProb) queryZone(a *Agent) *Msg {
	zone := p.current
	zone.shuffle()
	tried := []*IPv4{}

	for _, server := range zone.servers {
		if len(server.ips) == 0 {
			// TODO: ask IP first. will do this after AddrProb is done
		}
		if len(server.ips) == 0 {
			continue
		}
		for _, ip := range server.ips {
			if haveIP(tried, ip) {
				continue
			}
			tried = append(tried, ip)
			a.p.Print("//as",
				zone.zone.String(),
				fmt.Sprintf("@%s(%s)", server.name.String(), ip.String()))
			resp := a.Query(ip, p.n, p.t)
			if resp == nil {
				a.p.Print("//unreachable", server.name.String())
				continue
			}

			msg := resp.Msg
			rcode := msg.Flags & F_RCODEMASK
			if !(rcode == RCODE_OKAY || rcode == RCODE_NAMEERROR) {
				a.p.Print("//svrerror", server.name.String())
				continue
			}

			found, redirect := p.findAns(msg, a)
			if found {
				a.p.Print("//found")
				p.nextZone(nil)
				return msg
			} else {
				if redirect == nil {
					a.p.Print("//non-exist")
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

func (p *RecurProb) findAns(msg *Msg, a *Agent) (bool, *ZoneServers) {
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
		if !name.SubOf(p.current.zone) {
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
			a.p.Print("warning", "multiple subzones")
			continue
		}
		if rr.Class != IN || rr.Type != NS {
			panic("redirect record is wrong type")
		}
		nsData, ok := rr.Rdata.(*RdName)
		if !ok {
			panic("redirect record is not RdName")
		}

		nsName := nsData.name
		for _, s := range redirect.servers {
			if nsName.Equal(s.name) {
				continue rrloop
			}
		}

		ns := &NameServer{nsName, []*IPv4{}}
		msg.FilterINRR(func(rr *RR, seg int) bool {
			if rr.Type != A || !rr.Name.Equal(nsName) {
				return false
			}
			ipData, ok := rr.Rdata.(*RdIP)
			if !ok || ipData.ip == nil {
				return false
			}
			ns.ips = append(ns.ips, ipData.ip)
			return false // handled already
		})

		redirect.servers = append(redirect.servers, ns)
	}

	if len(redirect.servers) == 0 {
		panic("where are my redirect servers")
	}

	p.nscache.AddZone(redirect)

	return false, redirect
}

func (p *RecurProb) ExpandVia(a *Agent) {
	if p.start != nil {
		p.nextZone(p.start)
	} else {
		best := p.nscache.BestFor(p.n)
		if best == nil {
			// TODO: record no start zone error
			return
		}
		p.nextZone(best)
	}

	for p.current != nil {
		p.answer = p.queryZone(a)
	}
}

func (p *RecurProb) IndentSub() bool {
	return true
}
