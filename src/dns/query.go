package dns

import (
	"fmt"
	"io"
	"net"
	"os"
	"pson"
	"time"
)

// a packet handler listens on a local UDP port
// and takes input Queries, it will callback the input Queries' callback
// when the response is received or on timeout
// a connection only need to handle direct queries
type Conn struct {
	conn        net.PacketConn
	jobs        map[uint16]*queryJob
	sendQueue   chan *queryJob // scheduled queries
	recvQueue   chan *recvBuf  //
	closeSignal chan int
	recvClosed  chan int
	serveClosed chan int
	errlog      chan error
}

type queryJob struct {
	name     *Name
	t        uint16
	host     IPv4
	deadline time.Time
	signal   chan error
	resp     *Response
}

type ConnError struct {
	s string
}

func (e *ConnError) Error() string {
	return e.s
}

type ServeError struct {
	s string
	e error
}

func (e *ServeError) Error() string {
	return fmt.Sprintf("%s: %s", e.s, e.e)
}

func (c *Conn) logError(s string, e error) {
	c.errlog <- &ServeError{s, e}
}

type TimeoutError struct {
}

func (e *TimeoutError) Error() string {
	return "time out"
}

func (c *Conn) handleRecv(msg *Msg, addr net.Addr) error {
	switch udpa := addr.(type) {
	case *net.UDPAddr:
		ip := udpa.IP.To4()
		if ip == nil {
			return &ConnError{"host not ipv4"}
		}
		var ipv4 IPv4
		copy(ipv4[:], ip[:4])
		port := uint16(udpa.Port)

		job, b := c.jobs[msg.ID]
		if !b {
			return &ConnError{"no such id, time out already?"}
		}
		if !job.host.Equal(&ipv4) {
			return &ConnError{"recv from other hosts"}
		}

		job.resp = &Response{msg, ipv4, port, time.Now()}
		job.signal <- nil

		delete(c.jobs, msg.ID)
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
	cleanTicker := time.NewTicker(time.Second / 2) // check every half second
	idleTicker := time.NewTicker(time.Millisecond)

	for {
		didNothing := true
		// dispatch recv queue
		for len(c.recvQueue) > 0 {
			recv := <-c.recvQueue
			msg, err := FromWire(recv.buf)
			if err != nil {
				c.logError("parse", err)
			} else {
				err = c.handleRecv(msg, recv.addr)
				if err != nil {
					c.logError("handle", err)
				}
			}
			didNothing = false
		}

		if len(cleanTicker.C) > 0 { // time to clean time outs
			t := time.Now()
			toRemove := []uint16{}
			for id, job := range c.jobs {
				if t.After(job.deadline) {
					job.signal <- new(TimeoutError)
					toRemove = append(toRemove, id)
				}
			}

			for _, id := range toRemove {
				delete(c.jobs, id)
			}

			<-cleanTicker.C // sechedule next check
			didNothing = false
		}

		// send one if possible
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
				job.deadline = time.Now().Add(time.Second * 5)
				c.jobs[msg.ID] = job
			}

			didNothing = false
		}

		if len(c.closeSignal) > 0 {
			break
		}

		if didNothing {
			<-idleTicker.C
		}
	}

	cleanTicker.Stop()
	c.serveClosed <- 1
}

func (c *Conn) serveRecv() {
	wait := time.Second / 2
	buf := make([]byte, 512)
	for {
		deadline := time.Now().Add(wait)
		c.conn.SetReadDeadline(deadline)
		n, addr, err := c.conn.ReadFrom(buf)
		if err == nil {
			c.recvQueue <- &recvBuf{buf[:n], addr}
			buf = make([]byte, 512) // make a new one
		} else {
			if !err.(net.Error).Timeout() &&
				!err.(net.Error).Temporary() {
				c.logError("readFrom", err)
			}
		}

		if len(c.closeSignal) > 0 {
			break
		}
	}
	c.recvClosed <- 1
}

func (c *Conn) LogWith(f func(error)) {
	if c.errlog != nil {
		close(c.errlog)
	}

	c.errlog = make(chan error)
	go func() {
		for e := range c.errlog {
			f(e)
		}
	}()
}

func (c *Conn) LogToStderr() {
	c.LogWith(func(e error) {
		fmt.Fprintf(os.Stderr, "conn: %s\n", e)
	})
}

func (c *Conn) Start() error {
	c.sendQueue = make(chan *queryJob, 100)
	c.recvQueue = make(chan *recvBuf, 100)
	c.closeSignal = make(chan int, 1)
	c.recvClosed = make(chan int, 1)
	c.serveClosed = make(chan int, 1)
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return err
	}
	c.conn = conn
	go c.serveRecv() // receiving
	go c.serve()     // sending, time out and parsing

	return nil
}

// close gracefully
func (c *Conn) Stop() {
	if c.closeSignal == nil {
		return
	}

	c.closeSignal <- 0
	<-c.recvClosed
	<-c.serveClosed

	c.conn.Close()

	c.closeSignal = nil
}

func NewConn() (c *Conn, e error) {
	ret := new(Conn)
	ret.conn = nil
	ret.jobs = map[uint16]*queryJob{}

	return ret, nil
}

func (c *Conn) QueryHost(h *IPv4, n *Name, t uint16) (
	resp *Response, err error) {
	job := new(queryJob)
	job.name = n
	job.t = t
	job.host = *h
	job.signal = make(chan error, 1)
	job.deadline = time.Now()

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
			r, e := a.conn.QueryHost(&h, n, t)
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
