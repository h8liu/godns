package dns

import (
	"io"
)

type Client struct {
	conn  *Conn
	cache *NSCache
}

func NewClient() *Client {
	return &Client{NewConn(), DefNSCache}
}

func (c *Client) Solve(p ProbCase, log io.Writer) {
	solver := NewSolver(c.conn, log)
	solver.UseCache(c.cache)
	solver.Solve(p)
}

func (c *Client) Query(host *IPv4, name *Name, t uint16) (*Response, error) {
	re, err := c.conn.Query(host, name, t)
	return re, err
}
