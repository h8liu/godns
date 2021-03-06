package dns

import (
	"fmt"
	"io"
	"time"
)

const (
	_SOLVER_RETRY     = 3
	_SOLVER_MAX_DEPTH = 5
	_SOLVER_MAX_QUERY = 50
)

// the instruction set that a problem can use
type Solver interface {
	Query(host *IPv4, name *Name, t uint16) (resp *Response)
	SolveSub(p Prob) bool
	Log(args ...string)
	Cache(servers *Zone)
	QueryCache(zone *Name) *Zone
}

// a solver solves a problem recursively
// it serves as an execution engine for a problem
// and at the same time serves as a problem solver to the client
// it also records the message history through the solving proc
// a single solver instance can only be used for solving one problem
type solver struct {
	conn       *Conn
	p          *printer
	log        io.Writer
	signal     chan error
	cache      *NSCache
	rootProb   Prob
	checkpoint time.Time
	depth      int
	count      int
}

func newSolver(conn *Conn, log io.Writer) *solver {
	return &solver{
		conn:   conn,
		p:      newPrinter(),
		log:    log,
		signal: make(chan error, 1),
		cache:  TheCache,
	}
}

func (s *solver) UseCache(c *NSCache) {
	s.cache = c
}

func (s *solver) flushLog() {
	if s.log != nil {
		s.p.FlushTo(s.log)
	}
}

func durationStr(d time.Duration) string {
	ns := d.Nanoseconds()
	s := ns / 1000000000
	ms := (ns % 1000000000) / 1000000
	if ms == 0 && s == 0 {
		return fmt.Sprintf("+0")
	}
	if ms == 0 {
		return fmt.Sprintf("+%ds", s)
	}
	if s == 0 {
		return fmt.Sprintf("+%dms", ms)
	}
	return fmt.Sprintf("+%ds%dms", s, ms)
}

func (s *solver) lapse(t time.Time) time.Duration {
	ret := t.Sub(s.checkpoint)
	s.checkpoint = t
	return ret
}

func (s *solver) Query(h *IPv4, n *Name, t uint16) (resp *Response) {
	if s.count >= _SOLVER_MAX_QUERY {
		s.Log("err", fmt.Sprintf("too many queries (%d)", s.count))
		return nil // max count
	}
	s.count++

	for i := 0; i < _SOLVER_RETRY; i++ {
		s.Log("q", n.String(), TypeStr(t),
			fmt.Sprintf("@%s", h),
			durationStr(s.lapse(time.Now())))
		s.flushLog()
		s.conn.SendQuery(h, n, t,
			func(r *Response, e error) {
				resp = r
				s.signal <- e
			})
		err := <-s.signal
		if err == nil {
			s.p.PrintIndent("a", durationStr(s.lapse(resp.RecvTime)))
			resp.Msg.printTo(s.p)
			s.p.EndIndent()
			return
		}
		s.Log("err", err.Error(), durationStr(s.lapse(time.Now())))
	}

	return nil
}

func (s *solver) SolveSub(p Prob) bool {
	title := p.Title()
	s.Log(title...)

	if s.depth >= _SOLVER_MAX_DEPTH {
		s.Log("err", "too deep")
		return false
	}
	s.depth++
	s.p.Indent()
	p.ExpandVia(s) // solve the problem
	s.p.EndIndent()
	s.depth--

	return true
}

func (s *solver) Solve(p Prob) {
	if s.rootProb != nil {
		panic("agent consumed already")
	}
	s.count = 0
	s.depth = 0
	s.checkpoint = time.Now()
	s.rootProb = p
	s.SolveSub(p)
	s.flushLog()
}

func (s *solver) Log(args ...string) {
	s.p.Print(args...)
}

func (s *solver) Cache(zone *Zone) {
	s.cache.Add(zone)
}

func (s *solver) QueryCache(name *Name) *Zone {
	return s.cache.Query(name)
}
