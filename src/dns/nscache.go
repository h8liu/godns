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

type NSCache struct {
	cache     map[string]*cacheEntry
	lastClean time.Time
	syncLock  chan int
}

// the default nameserver cache
var DefNSCache *NSCache = NewNSCache()

// cache cleanup interval
const _CLEAN_INTERVAL = time.Hour / 2
const _DEFAULT_EXPIRE = time.Hour

func NewNSCache() *NSCache {
	ret := &NSCache{
		cache:     make(map[string]*cacheEntry),
		lastClean: time.Now(),
		syncLock:  make(chan int, 1),
	}
	ret.syncLock <- 0
	return ret
}

func (c *NSCache) lock() {
	<-c.syncLock
}

func (c *NSCache) unlock() {
	c.syncLock <- 0
}

func (c *NSCache) BestFor(name *Name) *ZoneServers {
	c.lock()
	defer c.unlock()

	now := time.Now()
	for name != nil {
		entry, found := c.cache[name.String()]
		if found && entry.expire.After(now) {
			return entry.s
		}
		name = name.Parent()
	}
	return nil
}

func (c *NSCache) Add(zs *ZoneServers) {
	zone := zs.Zone
	// servers := zs.Servers

	c.lock()
	defer c.unlock()

	zoneStr := zone.String()
	_, found := c.cache[zoneStr]
	if !found {

	} else {

	}
}