package dns

import (
	"encoding/binary"
	"fmt"
	"net"
)

var enc = binary.BigEndian

// IMPORTANT: should be treated as immutable
type IPv4 struct {
	ip [4]byte
}

func IPFromIP(ip net.IP) *IPv4 {
	ip4 := ip.To4()
	if ip4 == nil {
		return nil
	}
	ret := new(IPv4)
	copy(ret.ip[:], ip4[:4])
	return ret
}

func (ip *IPv4) IP() net.IP {
	return net.IPv4(ip.ip[0], ip.ip[1], ip.ip[2], ip.ip[3])
}

func (ip *IPv4) Bytes() []byte {
	ret := make([]byte, 4)
	copy(ret, ip.ip[:])
	return ret
}

func (ip *IPv4) Uint() uint32 {
	return enc.Uint32(ip.ip[:])
}

func IPFromBytes(bytes []byte) *IPv4 {
	if len(bytes) != 4 {
		return nil
	}
	ret := new(IPv4)
	copy(ret.ip[:], bytes)
	return ret
}

func (ip *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip.ip[0], ip.ip[1], ip.ip[2], ip.ip[3])
}

// will return nil on parse error
func ParseIP(s string) *IPv4 {
	nip := net.ParseIP(s)
	if nip == nil {
		return nil
	}
	nip = nip.To4()
	if nip == nil {
		return nil
	}

	ret := new(IPv4)
	copy(ret.ip[:], nip)
	return ret
}

func (ip *IPv4) Equal(other *IPv4) bool {
	for i, b := range ip.ip {
		if b != other.ip[i] {
			return false
		}
	}
	return true
}
