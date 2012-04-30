package dns

import (
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
type RecurAsker struct {
	n       *Name
	t       uint16
	nscache *NSCache
	start   *ZoneServers
	current *ZoneServers
	end     *ZoneServers
}

type ZoneServers struct {
	zone    *Name
	servers []NameServer
}

type NameServer struct {
	name *Name
	ips  []IPv4
}

func serverShuffle(servers []NameServer) []NameServer {
	n := len(servers)
	ret := make([]NameServer, n)
	order := rand.Perm(n)
	for i, ind := range order {
		ret[i] = servers[ind]
	}

	return ret
}

func (zs *ZoneServers) sortServers() {
	res := []NameServer{}
	nameOnly := []NameServer{}

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

type NSCache struct {
}

var globalNSCache *NSCache = NewNSCache()

func NewNSCache() *NSCache {
	// TODO
	return new(NSCache)
}

func (c *NSCache) BestFor(name *Name) *ZoneServers {
	// TODO
	return nil
}

func NewRecurAsker(name *Name, t uint16) *RecurAsker {
	ret := new(RecurAsker)
	ret.n = name
	ret.t = t
	ret.nscache = nil

	ret.UseGlobalCache()
	ret.StartFromRoot()

	return ret
}

func (a *RecurAsker) StartFromRoot() {
	a.start = nil
}

func (a *RecurAsker) StartWith(zone *Name, servers []NameServer) {
	a.start = &ZoneServers{zone, servers}
}

func (a *RecurAsker) UseCache(cache *NSCache) {
	a.nscache = cache
}

func (a *RecurAsker) UseGlobalCache() {
	a.nscache = globalNSCache
}

func (a *RecurAsker) UseNoCache() {
	a.nscache = nil
}

func (a *RecurAsker) name() string {
	return "rec"
}

func (a *RecurAsker) header() []string {
	return []string{a.n.String(), TypeStr(a.t)}
}

func haveIP(ipList []IPv4, ip IPv4) bool {
	for _, i := range ipList {
		if i.Equal(&ip) {
			return true
		}
	}
	return false
}

func (a *RecurAsker) askZone(agent *agent) {
	zone := a.current
	zone.sortServers()
	tried := []IPv4{}

	for _, server := range zone.servers {
		if len(server.ips) == 0 {
			// TODO: ask IP first
		}

		if len(server.ips) == 0 {
			continue
		}

		for _, ip := range server.ips {
			if haveIP(tried, ip) {
				continue
			}
			tried = append(tried, ip)
			resp := agent.netQuery(a.n, a.t, ip)
			if resp != nil { // not timeout
				// TODO: handle response
			}
		}
	}
}

func (a *RecurAsker) shoot(agent *agent) error {
	if a.start != nil {
		a.current = a.start
	} else {
		a.current = a.nscache.BestFor(a.n)
		if a.current == nil {
			return errors.New("no start zone")
		}
	}

	// TODO:
	for {
		break
	}

	return nil
}

// recursively query related records for a domain
type RecordAsker struct {
}

// recursively query an IP address for a domain
// will also chase down cnames
type IPAsker struct {
}
