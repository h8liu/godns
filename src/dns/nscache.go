package dns

import (
	"time"
)

type ZoneServers struct {
	Zone    *Name
	Servers []*NameServer
}

type NameServer struct {
	Name *Name
	Ips  []*IPv4
}

// the cache is two level map: zone -> server -> ip
// each server has an expiration date

type cacheEntry struct {
	s      *ZoneServers
	expire time.Time
}

type cacheRequest struct {
	newZone   *ZoneServers // nul if not an add
	queryZone *Name        // nul if is not an query
	queryChan chan *ZoneServers
}

type NSCache struct {
	cache     map[string]*cacheEntry
	lastClean time.Time
	requests  chan *cacheRequest
}

// the default nameserver cache
var TheCache *NSCache = NewNSCache()

func NewNSCache() *NSCache {
	ret := &NSCache{
		cache:     make(map[string]*cacheEntry),
		lastClean: time.Now(),
	}

	go ret.serve()

	return ret
}

func (c *NSCache) Close() {
	close(c.requests)
}

func (c *NSCache) Query(name *Name) *ZoneServers {
	queryChan := make(chan *ZoneServers)
	req := &cacheRequest{nil, name, queryChan}
	c.requests <- req
	return <-queryChan
}

func (c *NSCache) Add(zs *ZoneServers) {
	req := &cacheRequest{zs, nil, nil}
	c.requests <- req
}

// cache cleanup interval
const _CLEAN_INTERVAL = time.Hour / 2
const _DEFAULT_EXPIRE = time.Hour

func (c *NSCache) serve() {
	for req := range c.requests {
		if req.newZone != nil {
            c.serveAdd(req.newZone)
		}

		if req.queryZone != nil {
			if req.queryChan == nil {
				panic("req querychan empty")
			}
            req.queryChan <- c.serveQuery(req.queryZone)
		}
	}
}

func (c *NSCache) serveAdd(name *ZoneServers) {
    // TODO
}


func (c *NSCache) serveQuery(name *Name) *ZoneServers {
    // TODO
    return nil
}
