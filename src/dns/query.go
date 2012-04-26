package dns

import (
	"io"
	"pson"
	"time"
)

// a packet handler listens on a local UDP port
// and takes input Queries, it will callback the input Queries' callback
// when the response is received or on timeout
// a connection only need to handle direct queries
type Conn struct {
}

func (c *Conn) query(q *Ques, h IPv4) (*Response, error) {
	panic("not implemented")
	return nil, nil
}

func (c *Conn) Query(q Query, out io.Writer) {
	agent := &agent{pson.NewPrinter(), c, out}
	agent.query(q)
	agent.log.End()
	if agent.out != nil {
		agent.log.Flush(agent.out)
	}
}

// will record a tree structure of queries
// will only handle DirectQuery with conn
type agent struct {
	log  *pson.Printer
	conn *Conn
	out  io.Writer
}

func (a *agent) query(q Query) {
	a.log.PrintIndent(q.name(), q.header()...)
	q.run(a, a.log)
	a.log.EndIndent()
}

type Response struct {
	Msg  Msg
	Host IPv4
	Port uint16
	Time time.Time
}

func (a *agent) netQuery(q *Ques, hosts []IPv4) *Response {
	q.Pson(a.log) // question
	for _, h := range hosts {
		a.log.Print("ask", h.String())
		for i := 0; i < 3; i++ {
			if a.out != nil {
				a.log.Flush(a.out)
			}
			r, e := a.conn.query(q, h)
			if e == nil {
				a.log.Print("recv", r.Time.String())
				r.Msg.Pson(a.log)
				return r
			}
			a.log.Print("error", e.Error())
		}
	}
	return nil
}

type Query interface {
	run(a *agent, log *pson.Printer)
	name() string
	header() []string
}

// recursively query related records for a domain
type RecordQuery struct {
}

// recursively query an IP address for a domain
type IPQuery struct {
}

// recursively query a question through a bunch of servers
type RecurQuery struct {
}
