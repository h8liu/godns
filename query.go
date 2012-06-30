package dns

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// maintains a dns connection for dns queries
type Conn struct {
	conn        net.PacketConn
	jobs        map[uint16]*queryJob
	sendQueue   chan *queryJob // scheduled queries
	recvQueue   chan *recvBuf  // received packets
	errlog      chan error
	closeSignal chan int
	recvClosed  chan int
	serveClosed chan int
	started     bool
}

// an internal async query job
type queryJob struct {
	name     *Name
	t        uint16
	host     *IPv4
	deadline time.Time
	callback QueryCallback
}

// a received packet
type recvBuf struct {
	buf  []byte
	addr net.Addr
}

// a parsed response
type Response struct {
	Msg  *Msg
	Host *IPv4
	Port uint16
	Time time.Time
}

// encapsulate background thread error
type ConnError struct {
	s string
	e error
}

func (e *ConnError) Error() string {
	return fmt.Sprintf("%s: %s", e.s, e.e)
}

func (c *Conn) logError(s string, e error) {
	c.errlog <- &ConnError{s, e}
}

// time out error
type TimeoutError struct {
}

func (e *TimeoutError) Error() string {
	return "time out"
}

func (c *Conn) handleRecv(msg *Msg, addr net.Addr) error {
	switch udpa := addr.(type) {
	case *net.UDPAddr:
		ip := IPFromIP(udpa.IP)
		if ip == nil {
			return errors.New("host not ipv4")
		}
		port := uint16(udpa.Port)

		job, b := c.jobs[msg.ID]
		if !b {
			return errors.New("no such id, time out already?")
		}
		if !job.host.Equal(ip) {
			return errors.New("recv from other hosts")
		}

		resp := &Response{msg, ip, port, time.Now()}
		job.callback(resp, nil)

		delete(c.jobs, msg.ID)
	default:
		return errors.New("addr not UDP")
	}
	return nil
}

func (c *Conn) serve() {
	// check every half second
	cleanTicker := time.NewTicker(time.Second / 2)
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
					job.callback(nil, new(TimeoutError))
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
				msg.RollID()
				_, b = c.jobs[msg.ID]
			}
			buf, err := msg.Wire()
			if err == nil {
				ip := job.host.ToIP()
				addr := &net.UDPAddr{ip, 53}
				_, err = c.conn.WriteTo(buf, addr)
			}

			if err != nil {
				job.callback(nil, err)
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

func (c *Conn) recv() {
	wait := time.Millisecond
	buf := make([]byte, 512)
	for {
		deadline := time.Now().Add(wait)
		c.conn.SetReadDeadline(deadline)
		n, addr, err := c.conn.ReadFrom(buf)
		if err == nil {
			c.recvQueue <- &recvBuf{buf[:n], addr}
			buf = make([]byte, 512) // make a new one
		} else {
			if nerr, b := err.(net.Error); b {
				if !nerr.Timeout() &&
					!nerr.Temporary() {
					c.logError("readFrom1", nerr)
				}
			} else {
				c.logError("readFrom2", err)
			}
		}

		if len(c.closeSignal) > 0 {
			break
		}
	}

	c.recvClosed <- 1
}

type Logger func(error)

func LogToStderr(e error) {
	fmt.Fprintf(os.Stderr, "%s\n", e)
}

func (c *Conn) LogWith(log Logger) {
	if c.errlog != nil {
		close(c.errlog)
	}

	c.errlog = make(chan error)
	go func() {
		for e := range c.errlog {
			log(e)
		}
	}()
}

func (c *Conn) start() error {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return err
	}
	c.conn = conn

	c.sendQueue = make(chan *queryJob, 100)
	c.recvQueue = make(chan *recvBuf, 100)
	c.closeSignal = make(chan int, 1)
	c.recvClosed = make(chan int, 1)
	c.serveClosed = make(chan int, 1)
	c.started = true

	go c.recv()  // receiving
	go c.serve() // sending, time out and parsing

	return nil
}

// close gracefully
func (c *Conn) Close() {
	if c.closeSignal == nil {
		// already closed
		return
	}

	c.closeSignal <- 0
	<-c.recvClosed
	<-c.serveClosed

	c.conn.Close()

	c.closeSignal = nil
}

// creates a connection
func NewConn() (c *Conn, e error) {
	ret := new(Conn)
	ret.conn = nil
	ret.jobs = map[uint16]*queryJob{}
	ret.started = false
	ret.LogWith(LogToStderr)

	return ret, nil
}

func (c *Conn) ensureStarted() error {
	if !c.started {
		return c.start()
	}
	return nil
}

type QueryCallback func(*Response, error)

func (c *Conn) SendQuery(h *IPv4, n *Name, t uint16,
	callback QueryCallback) {
	err := c.ensureStarted()
	if err != nil {
		callback(nil, err)
		return
	}

	job := &queryJob{name: n, t: t, host: h, callback: callback}

	c.sendQueue <- job
}

func (c *Conn) Query(h *IPv4, n *Name, t uint16) (
	resp *Response, err error) {

	signal := make(chan error, 1)
	c.SendQuery(h, n, t, func(r *Response, e error) {
		resp = r
		signal <- e
	})

	err = <-signal
	if err != nil {
		return nil, err
	}
	return resp, nil
}
