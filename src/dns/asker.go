package dns

import (
	"pson"
)

type Asker interface {
	shoot(a *agent, log *pson.Printer)
	name() string
	header() []string
}

// recursively query related records for a domain
type RecordAsker struct {
}

// recursively query an IP address for a domain
type IPAsker struct {
}

// recursively query a question through a bunch of servers
type RecurAsker struct {
}

func NewRecurAsker(name *Name, t uint16) *RecurAsker {
	ret := new(RecurAsker)

	return ret
}

func (q *RecurAsker) BeginWith(zone *Name, ns []IPv4) {

}

func (q *RecurAsker) shoot(a *agent, log *pson.Printer) {

}

func (q *RecurAsker) name() string {
	return "rec"
}

func (q *RecurAsker) header() []string {
	return nil
}
