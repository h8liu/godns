package dns

import "math/rand"

type Zone struct {
	name    *Name
	servers map[string]*NameServer
	ips     map[uint32]*Name
}

type NameServer struct {
	Name *Name
	IPs  []*IPv4
}

func NewZone(name *Name) *Zone {
	return &Zone{
		name,
		make(map[string]*NameServer),
		make(map[uint32]*Name),
	}
}

func (old *Zone) Copy() *Zone {
	ret := NewZone(old.name)
	for name, server := range old.servers {
		// only need to duplicate the map
		ret.servers[name] = &NameServer{
			server.Name,
			server.IPs, // this is okay since the content will not be changed
			// Add to the origin will replace the slice pointer
		}
	}

	for i, name := range old.ips {
		ret.ips[i] = name
	}

	return ret
}

func (self *Zone) Add(serverName *Name, ips ...*IPv4) {
	toAdd := make([]*IPv4, 0, len(ips))

	for _, ip := range ips {
		i := ip.Uint()
		if self.ips[i] == nil {
			self.ips[i] = serverName
			toAdd = append(toAdd, ip)
		}
	}

	nameStr := serverName.String()
	s := self.servers[nameStr]
	if s != nil {
		if len(toAdd) > 0 {
			s.IPs = append(s.IPs, toAdd...)
		}
		return
	}

	self.servers[nameStr] = &NameServer{
		serverName,
		toAdd,
	}
}

func (self *Zone) AddName(serverName *Name) {
	nameStr := serverName.String()
	s := self.servers[nameStr]
	if s != nil {
		return
	}

	self.servers[nameStr] = &NameServer{
		serverName,
		[]*IPv4{},
	}
}

func randOrder(servers []*NameServer) []*NameServer {
	n := len(servers)
	ret := make([]*NameServer, n)
	order := rand.Perm(n)
	for i, ind := range order {
		ret[i] = servers[ind]
	}

	return ret
}

func shuffle(servers []*NameServer) []*NameServer {
	ret := make([]*NameServer, 0, len(servers))
	nameOnly := make([]*NameServer, 0, len(servers))

	for _, ns := range servers {
		if len(ns.IPs) == 0 {
			nameOnly = append(nameOnly, ns)
		} else {
			ret = append(ret, ns)
		}
	}

	ret = randOrder(ret)
	ret = append(ret, (randOrder(nameOnly))...)

	return ret
}

func (self *Zone) Name() *Name {
	return self.name
}

func (self *Zone) Prepare() []*NameServer {
	servers := shuffle(self.List())
	return servers
}

func (self *Zone) List() []*NameServer {
	servers := make([]*NameServer, 0, len(self.servers))

	for _, server := range self.servers {
		servers = append(servers, server)
	}

	return servers
}
