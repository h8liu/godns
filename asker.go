package dns

import (
	"./pson"
	"errors"
	"math/rand"
)

type Asker interface {
	shoot(a *agent) error
	name() string
	header() []string
}

// recursively query a question through a bunch of servers
// only focus on one single record type
type Recursive struct {
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

func serverShuffle(servers []*NameServer) []*NameServer {
	n := len(servers)
	ret := make([]*NameServer, n)
	order := rand.Perm(n)
	for i, ind := range order {
		ret[i] = servers[ind]
	}

	return ret
}

func (zs *ZoneServers) sortServers() {
	res := []*NameServer{}
	nameOnly := []*NameServer{}

	for _, ns := range zs.servers {
		if len(ns.ips) == 0 {
			nameOnly = append(nameOnly, ns)
		} else {
			res = append(res, ns)
		}
	}

	res = serverShuffle(res)
	nameOnly = serverShuffle(nameOnly)
	for _, ns := range nameOnly {
		res = append(res, ns)
	}
	zs.servers = res
}

func NewRecursive(name *Name, t uint16) *Recursive {
	ret := new(Recursive)
	ret.n = name
	ret.t = t
	ret.nscache = nil
	ret.answer = nil

	ret.UseGlobalCache()
	ret.StartFromRoot()

	return ret
}

func (a *Recursive) StartFromRoot() {
	a.start = nil
}

func (a *Recursive) StartWith(zone *Name, servers []*NameServer) {
	a.start = &ZoneServers{zone, servers}
}

func (a *Recursive) UseNoCache() {
	a.UseCache(nil)
}

func (a *Recursive) UseCache(cache *NSCache) {
	a.nscache = cache
}

func (a *Recursive) UseGlobalCache() {
	a.UseCache(globalNSCache)
}

func (a *Recursive) name() string {
	return "rec"
}

func (a *Recursive) header() []string {
	return []string{a.n.String(), TypeStr(a.t)}
}

func haveIP(ipList []*IPv4, ip *IPv4) bool {
	for _, i := range ipList {
		if i.Equal(ip) {
			return true
		}
	}
	return false
}

func (a *Recursive) nextZone(zs *ZoneServers) {
	if zs != nil {
		a.last = zs
	}
	a.current = zs
}

func (a *Recursive) askZone(agent *agent) *Msg {
	zone := a.current
	zone.sortServers()
	tried := []*IPv4{}

	for _, server := range zone.servers {
		if len(server.ips) == 0 {
			// TODO: ask IP first
		}
		if len(server.ips) == 0 {
			continue
		}
		agent.log.Print("use", server.name.String())
		for _, ip := range server.ips {
			if haveIP(tried, ip) {
				continue
			}
			tried = append(tried, ip)
			resp := agent.netQuery(a.n, a.t, ip)
			if resp == nil {
				continue
			}

			msg := resp.Msg
			rcode := msg.Flags & F_RCODEMASK
			// we only trust okay and name errors
			// for other error codes, we will simply treat them
			// as time outs
			if !(rcode == RCODE_OKAY || rcode == RCODE_NAMEERROR) {
				continue
			}

			found, redirect := a.findAns(msg, agent.log)
			if found {
				a.nextZone(nil)
				return msg
			} else {
				a.nextZone(redirect)
				return nil
			}
		}
	}

	// got nothing, so set next zone to nil
	a.nextZone(nil)
	return nil
}

func (a *Recursive) findAns(msg *Msg, log *pson.Printer) (bool, *ZoneServers) {
	// look for answer
	rrs := msg.FilterINRR(func(rr *RR, seg string) bool {
		if !rr.Name.Equal(a.n) {
			return false
		}
		return a.t == rr.Type || (a.t == A && rr.Type == CNAME)
	})
	if len(rrs) > 0 {
		return true, nil
	}

	// look for redirect name servers
	rrs = msg.FilterINRR(func(rr *RR, seg string) bool {
		if rr.Type != NS {
			return false
		}
		name := rr.Name
		if !name.SubOf(a.current.zone) {
			return false
		}
		if !name.Equal(a.n) && !name.ParentOf(a.n) {
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
		if rr.Name.Equal(subzone) {
			log.Print("weird", "multiple subzones")
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
		msg.FilterINRR(func(rr *RR, seg string) bool {
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

	a.nscache.AddZone(redirect)

	return false, redirect
}

func (a *Recursive) shoot(agent *agent) error {
	if a.start != nil {
		a.nextZone(a.start)
	} else {
		best := a.nscache.BestFor(a.n)
		if best == nil {
			return errors.New("no start zone")
		}
		a.nextZone(best)
	}

	for a.current != nil {
		a.answer = a.askZone(agent)
	}
	return nil
}

// recursively query an IP address for a domain
// will also chase down cnames
// TODO
type IPAsker struct {
}

// recursively query related records for a domain
// TODO
type RecordAsker struct {
}
