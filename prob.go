package dns

import (
    "math/rand"
)

type Prob interface {
	IndentSub() bool
	ExpandVia(a *Agent)
	Title() (name string, meta []string)
}

type ZoneServer struct {
    name *Name
    ips []*IPv4
}

type ZoneProb struct {
    n *Name
    t uint16
    zone *Name
    servers []*ZoneServer
}

func shuffleServers(servers []*ZoneServer) []*ZoneServer {
    n := len(servers)
    ret := make([]*ZoneServer, n)
    for i, ind := range rand.Perm(n) {
        ret[i] = servers[ind]
    }
    return ret
}

func sortServers(servers []*ZoneServer) []*ZoneServer {
    ret := []*ZoneServer{}
    nameOnly := []*ZoneServer{}
    for _, ns := range servers {
        if len(ns.ips) == 0 {
            nameOnly = append(nameOnly, ns)
        } else {
            ret = append(ret, ns)
        }
    }

    ret = shuffleServers(ret)
    nameOnly = shuffleServers(nameOnly)
    for _, ns := range nameOnly {
        ret = append(ret, ns)
    }

    return ret
}

func NewZoneProb(name *Name, t uint16, zone *Name, servers []*ZoneServer) *ZoneProb{
    return &ZoneProb{n:name, t:t, zone:zone, servers:servers}
}

/*
func (zp *ZoneProb) ExpandVia(a *Agent) {
    servers := sortServers(zp.servers)
    
    tried := []*IPv4{}

    for _, server := range zone.servers {
        if len(server.ips) == 0 {
            // TODO: ask IP first
        }

        if len(server.ips) == 0 {
            continue
        }
        
        a.p.Print("use", server.name.String())
        for _, ip := range server.ips {
            if haveIP(tried, ip) {
                continue
            }
            tried = append(tried, ip)
            resp := a.Query(ip, a.n, a.t)
            if resp == nil {
                continue
            }

            msg := resp.Msg
            rcode := msg.Flags & F_RCODEMASK
            if !(rcode == RCODE_OKAY || rcode == RCODE_NAMEERROR) {
                continue
            }

            found, redirect := zp.findAns(msg, agent.log)
            if found {
                a.

}
*/
