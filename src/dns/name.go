package dns

import (
	"fmt"
	"strings"
)

// immutable 
type Name struct {
	labels []string
}

type nameError struct {
	name string
	s    string
}

func (e *nameError) Error() string {
	return "name '" + e.name + "': " + e.s
}

func (n *Name) Equal(other *Name) bool {
	if other == nil {
		return false
	}
	if n == nil {
		return false
	}
	if len(n.labels) != len(other.labels) {
		return false
	}
	for i, label := range n.labels {
		if label != other.labels[i] {
			return false
		}
	}
	return true
}

func (n *Name) String() string {
	if n == nil {
		fmt.Println("nil lables")
	}
	if len(n.labels) > 0 {
		return strings.Join(n.labels, ".")
	}
	return "."
}

// for programming use, will panic on fail
// use with caution
func MakeName(s string) *Name {
	ret, err := NewName(s)
	if err != nil {
		panic(fmt.Sprintf("makeName failed: %s", s))
	}
	return ret
}

func NewName(s string) (ret *Name, e error) {
	if len(s) == 0 {
		return nil, &nameError{s, "empty name"}
	}

	if len(s) > 255 {
		return nil, &nameError{s, "name too long"}
	}

	if s[len(s)-1] != '.' {
		s += "."
	}

	labels := make([]string, 0)

	last := byte('.')
	ok := false
	partlen := 0
	start := 0

	if s != "." {
		for i := 0; i < len(s); i++ {
			c := s[i]
			switch {
			default:
				return nil, &nameError{s, "special characters"}
			case 'a' <= c && c <= 'z':
				fallthrough
			case 'A' <= c && c <= 'Z':
				c = c - 'A' + 'a'
				fallthrough
			case c == '_':
				ok = true
				partlen++
			case '0' <= c && c <= '9':
				partlen++
			case c == '-':
				if last == '.' {
					return nil, &nameError{s, "dash before dot"}
				}
				partlen++
			case c == '.':
				if last == '.' {
					return nil, &nameError{s, "consecutive dot"}
				}
				if last == '-' {
					return nil, &nameError{s, "dash after dot"}
				}
				if partlen > 63 {
					return nil, &nameError{s, "label too long"}
				}
				if partlen == 0 {
					return nil, &nameError{s, "start with dot, empty label"}
				}
				partlen = 0
				labels = append(labels, string(s[start:i]))
				start = i + 1
			}
			last = c
		}

		if !ok {
			return nil, &nameError{s, "all numbers, maybe IP"}
		}
	}

	ret = &Name{labels}
	return ret, nil
}

func (n *Name) SubOf(other *Name) bool {
	if len(other.labels) == 0 {
		return true
	}
	if len(n.labels) <= len(other.labels) {
		return false
	}
	subs := n.labels[len(n.labels)-len(other.labels):]
	for i, lab := range subs {
		if lab != other.labels[i] {
			return false
		}
	}
	return true
}

func (n *Name) ParentOf(other *Name) bool {
	return other.SubOf(n)
}

func (n *Name) IsRoot() bool {
	return len(n.labels) == 0
}

func (n *Name) Parent() *Name {
	if n.IsRoot() {
		return nil
	}
	labels := make([]string, len(n.labels)-1)
	copy(labels, n.labels[1:])
	return &Name{labels}
}
