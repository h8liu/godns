package dns

import (
	"fmt"
	"net"
)

// should be treated as immutable
type IPv4 [4]byte

func (ip *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip[0], ip[1], ip[2], ip[3])
}

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
	copy(ret[:], nip)
	return ret
}

func (ip *IPv4) Equal(other *IPv4) bool {
	for i, b := range ip {
		if b != other[i] {
			return false
		}
	}
	return true
}
