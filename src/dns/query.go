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
	conn      net.PacketConn
	jobs      map[uint16]*queryJob
	sendQueue chan *queryJob
	recvQueue chan *recvBuf
	idRecycle chan uint16
}

type queryJob struct {
	name   *Name
	t      uint16
	host   IPv4
	sent   time.Time
	signal chan error
	resp   *Response
}

type ConnError struct {
	s string
}

func (e *ConnError) Error() string {
	return e.s
}

func (c *Conn) handle(msg *Msg, addr net.Addr) error {
	switch udpa := addr.(type) {
	case *net.UDPAddr:
		ip := udpa.IP.To4()
		if ip == nil {
			return &ConnError{"host not ipv4"}
		}
		var ipv4 IPv4
		copy(ipv4[:], ip[:4])
		port := uint16(udpa.Port)
		resp := &Response{msg, ipv4, port, time.Now()}
		if resp != nil {
		} // just to disable the compile error
		// TODO: channel response back
	default:
		return &ConnError{"addr not UDP"}
	}
	return nil
}

type recvBuf struct {
	buf  []byte
	addr net.Addr
}

func (c *Conn) serve() {
	for {
		// dispatch recv queue
		for len(c.recvQueue) > 0 {
			recv := <-c.recvQueue
			msg, err := FromWire(recv.buf)
			if err != nil {
				// TODO: log parsing error
			} else {
				err = c.handle(msg, recv.addr)
				if err != nil {
					// TODO: log handle error
				}
			}
		}

		// TODO: check all time outs if needed

		// send jobs
		if len(c.sendQueue) > 0 {
			job := <-c.sendQueue
			msg := NewQuesMsg(job.name, job.t)
			_, b := c.jobs[msg.ID]
			for b {
				msg.RandID()
				_, b = c.jobs[msg.ID]
			}
			buf, err := msg.ToWire()
			if err == nil {
				ip := net.IPv4(job.host[0], job.host[1],
					job.host[2], job.host[3])
				addr := &net.UDPAddr{ip, 53}
				_, err = c.conn.WriteTo(buf, addr)
			}

			if err != nil {
				job.signal <- err
			} else {
				// send succeed, waiting now
				job.sent = time.Now()
				c.jobs[msg.ID] = job
			}
		}

		// TODO: check closing
	}
}

func (c *Conn) serveRecv() {
	buf := make([]byte, 512)
	for {
		deadline := time.Now().Add(time.Second / 2)
		c.conn.SetReadDeadline(deadline)
		n, addr, err := c.conn.ReadFrom(buf)
		if err != nil {
			out := make([]byte, n)
			copy(out, buf[:n])
			c.recvQueue <- &recvBuf{out, addr}
		} else {
			// TODO: log network error
		}

		// TODO: check closing
	}
}

func (c *Conn) Start() error {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return err
	}
	c.conn = conn
	go c.serveRecv() // receiving
	go c.serve()     // sending, time out and parsing

	return nil
}

func NewConn() (c *Conn, e error) {
	ret := new(Conn)
	ret.conn = nil
	ret.jobs = map[uint16]*queryJob{}
	ret.sendQueue = make(chan *queryJob, 100)
	ret.recvQueue = make(chan *recvBuf, 100)
	ret.idRecycle = make(chan uint16, 500)

	return ret, nil
}

func (c *Conn) query(n *Name, t uint16, h IPv4) (resp *Response, err error) {
	job := new(queryJob)
	job.name = n
	job.t = t
	job.host = h
	job.signal = make(chan error, 1)
	job.sent = time.Now()

	c.sendQueue <- job
	err = <-job.signal
	if err != nil {
		return nil, err
	}

	return job.resp, nil
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
	Msg  *Msg
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
