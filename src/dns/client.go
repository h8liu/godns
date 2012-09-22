package dns

import (
	"io"
)

// client is a synchronous helper for solving simple problems
// it will create a connection automatically
type Client struct {
	conn  *Conn
	cache *NSCache
}

func NewClient() *Client {
	return &Client{NewConn(), TheCache}
}

func (c *Client) Solve(p Prob, logTo io.Writer) {
	solver := newSolver(c.conn, logTo)
	solver.UseCache(c.cache)
	solver.Solve(p)
}

func (c *Client) RecurQuery(n *Name, t uint16, logTo io.Writer) *ProbRecur {
	solver := newSolver(c.conn, logTo)
	solver.UseCache(c.cache)
	recur := NewProbRecur(n, t)
	solver.Solve(recur)
	return recur
}

func (c *Client) Query(host *IPv4, name *Name, t uint16) (*Response, error) {
	re, err := c.conn.Query(host, name, t)
	return re, err
}
