package dns

import (
	"io"
	"net"
	"pson"
	"time"
)

// a packet handler listens on a local UDP port
// and takes input Queries, it will callback the input Queries' callback
// when the response is received or on timeout
// a connection only need to handle direct queries
type Conn struct {
	conn net.PacketConn
}

func (c *Conn) listen() {
	buf := make([]byte, 512)
	for {
		deadline := time.Now().Add(time.Second / 2)
		c.conn.SetReadDeadline(deadline)
		n, _, err := c.conn.ReadFrom(buf)
		if err != nil {
			msgbuf := buf[:n]
			_, err = FromWire(msgbuf)
			if err != nil {
				// TODO: log parsing error
			} else {
				// TODO: queue message back
				// search for the id
				// and call back with receiving
			}
		} else {
			// TODO: log network error
		}
		// TODO: check all time outs if needed

		// TODO: check if it is requested to close
	}
}

func NewConn() (c *Conn, e error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	ret := &Conn{conn}
	go ret.listen()
	return ret, nil
}

func (c *Conn) query(n *Name, t uint16, h IPv4) (*Response, error) {
	panic("not implemented")
	// allocate an id
	// make a query message
	// pack it
	// send it out
	// wait for reply on a channel

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

func (a *agent) netQuery(n *Name, t uint16, hosts []IPv4) *Response {
	a.log.Print("q", n.String(), TypeStr(t))
	for _, h := range hosts {
		a.log.Print("ask", h.String())
		for i := 0; i < 3; i++ {
			if a.out != nil {
				a.log.Flush(a.out)
			}
			r, e := a.conn.query(n, t, h)
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
