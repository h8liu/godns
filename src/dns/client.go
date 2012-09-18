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
	return &Client{NewConn(), DefNSCache}
}

func (c *Client) Solve(p Prob, log io.Writer) {
	solver := newSolver(c.conn, log)
	solver.UseCache(c.cache)
	solver.Solve(p)
}

func (c *Client) Query(host *IPv4, name *Name, t uint16) (*Response, error) {
	re, err := c.conn.Query(host, name, t)
	return re, err
}
