package dns

import (
	"strings"
	"fmt"
)

// immutable 
type Name struct {
	labels []string
}

type NameError struct {
	name string
	s    string
}

func (e *NameError) Error() string {
	return "name '" + e.name + "': " + e.s
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

func NewName(s string) (ret *Name, e error) {
	if len(s) == 0 {
		return nil, &NameError{s, "empty name"}
	}

	if len(s) > 255 {
		return nil, &NameError{s, "name too long"}
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
				return nil, &NameError{s, "special characters"}
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
					return nil, &NameError{s, "dash before dot"}
				}
				partlen++
			case c == '.':
				if last == '.' {
					return nil, &NameError{s, "consecutive dot"}
				}
				if last == '-' {
					return nil, &NameError{s, "dash after dot"}
				}
				if partlen > 63 {
					return nil, &NameError{s, "label too long"}
				}
				if partlen == 0 {
					return nil, &NameError{s, "start with dot, empty label"}
				}
				partlen = 0
				labels = append(labels, string(s[start:i]))
				start = i + 1
			}
			last = c
		}

		if !ok {
			return nil, &NameError{s, "all numbers, maybe IP"}
		}
	}

	ret = &Name{labels}
	return ret, nil
}
