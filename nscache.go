package dns

type NSCache struct {
	cache map[string]*ZoneServers
}

var DefNSCache *NSCache = NewNSCache()

func NewNSCache() *NSCache {
    ret := new(NSCache)
    ret.cache = make(map[string]*ZoneServers)
    return ret
}

func (c *NSCache) BestFor(name *Name) *ZoneServers {
	// TODO
	return nil
}

func (c *NSCache) AddZone(zs *ZoneServers) {
	c.AddServer(zs.zone, zs.servers...)
}

func (c *NSCache) AddServer(zone *Name, servers ...*NameServer) {

}
