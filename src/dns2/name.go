package dns2

import (
	"strings"
)

func (n *Name) String() string {
	return strings.Join(n.label, ".")
}
