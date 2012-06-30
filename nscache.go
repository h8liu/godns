package dns

type NSCache struct {
	cache map[string]*ZoneServers
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

func (c *NSCache) AddZone(zs *ZoneServers) {
	c.AddServer(zs.zone, zs.servers...)
}

func (c *NSCache) AddServer(zone *Name, servers ...*NameServer) {

}
