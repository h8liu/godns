package dns

import (
	"time"
)

// the cache is two level map: zone -> server -> ip
// each server has an expiration date

type cacheEntry struct {
	zone      *Zone
    ipnames    map[uint32] *Name
	expire time.Time
}

type cacheRequest struct {
	newZone    *Zone // nul if not an add
	queryZone  *Name        // nul if is not an query
	queryReply chan *Zone
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
        requests: make(chan *cacheRequest),
	}

	go ret.serve()

	return ret
}

func (c *NSCache) Close() {
	close(c.requests)
}

func (c *NSCache) Query(name *Name) *Zone {
	queryReply := make(chan *Zone)
	req := &cacheRequest{nil, name, queryReply}
	c.requests <- req
	return <-queryReply
}

func (c *NSCache) Add(zs *Zone) {
	req := &cacheRequest{zs, nil, nil}
	c.requests <- req
}

// cache cleanup interval
const _CLEAN_INTERVAL = time.Hour / 4
const _DEFAULT_EXPIRE = time.Hour

func (c *NSCache) serve() {
	for req := range c.requests {
		if req.newZone != nil {
			c.serveAdd(req.newZone)
		}

		if req.queryZone != nil {
			if req.queryReply == nil {
				panic("req queryReply is nil")
			}
			req.queryReply <- c.serveQuery(req.queryZone)
		}
	}
}

func NewEntry(servers *Zone) *cacheEntry {
    ipnames := make(map[uint32] *Name)
    s := NewZone(servers.Name())
    list := servers.List()

    for _, server := range list {
        for _, ip := range server.IPs {
            i := ip.Uint()
            _, has := ipnames[i]
            if has {
                continue
            }

            // now add this ip address to list
            ipnames[i] = server.Name
            s.Add(server.Name, ip)
        }
    }

    if len(ipnames) == 0 {
        return nil // nothing to add
    }

    return &cacheEntry{
        s,
        ipnames,
        time.Now().Add(_DEFAULT_EXPIRE),
    }
}

func (old *cacheEntry) Copy() *cacheEntry {
    ret := &cacheEntry{
        old.zone.Copy(),
        make(map[uint32] *Name),
        old.expire,
    }

    for i, name := range old.ipnames {
        ret.ipnames[i] = name
    }
    return ret
}

func (e *cacheEntry) add(servers []*NameServer) (changed bool) {
    for _, server := range servers {
        for _, ip := range server.IPs {
            i := ip.Uint()
            if e.ipnames[i] != nil {
                continue
            }
            
            e.ipnames[i] = server.Name
            e.zone.Add(server.Name, ip)
            changed = true       
        }
    }
    return
}

func (c *NSCache) serveAdd(zone *Zone) {
    zoneStr := zone.Name().String()
    curEntry := c.cache[zoneStr]
    
    if curEntry == nil { 
        entry := NewEntry(zone)
        if entry == nil {
            return
        }
        c.cache[zoneStr] = entry
        return
    } 
 
    newEntry := curEntry.Copy()
    if newEntry.add(zone.List()) {
        // entry changed, swap in the new one
        c.cache[zoneStr] = newEntry
    }
}

func (c *NSCache) serveQuery(name *Name) *Zone {
    entry := c.cache[name.String()]
    if entry != nil {
        return entry.zone
    }
    return nil
}
