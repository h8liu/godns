package dns

import (
	"./pson"
	"fmt"
	"io"
)

type Prob interface {
	IndentSub() bool
	ExpandVia(a *Agent)
	Title() (name string, meta []string)
}

const (
	AGENT_RETRY = 3
)

type Agent struct {
	conn   *Conn
	p      *pson.Printer
	log    io.Writer
	signal chan error
}

func NewAgent(conn *Conn, log io.Writer) *Agent {
	return &Agent{conn: conn, p: pson.NewPrinter(),
		log: log, signal: make(chan error, 1)}
}

func (a *Agent) flushLog() {
	if a.log != nil {
		a.p.FlushTo(a.log)
	}
}

func (a *Agent) Query(h *IPv4, n *Name, t uint16) (resp *Response) {
	a.p.Print("q", n.String(), TypeStr(t),
		fmt.Sprintf("@%s", h))
	for i := 0; i < AGENT_RETRY; i++ {
		a.flushLog()
		a.conn.SendQuery(h, n, t,
			func(r *Response, e error) {
				resp = r
				a.signal <- e
			})
		err := <-a.signal
		if err == nil {
			a.p.Print("recv", resp.Time.String())
			resp.Msg.PsonTo(a.p)
			return
		}
		a.p.Print("err", err.Error())
	}

	return nil
}

func (a *Agent) Solve(p Prob) {
	name, meta := p.Title()
	if meta == nil {
		a.p.Print(name)
	} else {
		a.p.Print(name, meta...)
	}

	indent := p.IndentSub()

	if indent {
		a.p.Indent()
	}

	p.ExpandVia(a) // solve the problem

	if indent {
		a.p.EndIndent()
	}
}
