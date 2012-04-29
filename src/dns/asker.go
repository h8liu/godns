package dns

import (
	"errors"
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
	servers []IPv4
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

func (a *RecurAsker) StartWith(zone *Name, servers []IPv4) {
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

func (a *RecurAsker) askZone() {

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
