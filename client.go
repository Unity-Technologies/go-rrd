package rrd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultPort is the default rrdcached port.
	DefaultPort = 42217
)

var (
	respRe = regexp.MustCompile(`^(-?\d+)\s+(.*)$`)

	// DefaultTimeout is the default read / write / dial timeout for Clients.
	DefaultTimeout = time.Second * 10
)

// Client is a rrdcached client.
type Client struct {
	conn    net.Conn
	addr    string
	network string
	timeout time.Duration
	scanner *bufio.Scanner
}

// Timeout sets read / write / dial timeout for a rrdcached Client.
func Timeout(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.timeout = timeout
		return nil
	}
}

// Unix sets the client to use a unix socket.
func Unix(c *Client) error {
	c.network = "unix"
	return nil
}

// NewClient returns a new rrdcached client connected to addr.
// By default addr is treated as a TCP address to use UNIX sockets pass Unix as an option.
// If addr for a TCP address doesn't include a port the DefaultPort will be used.
func NewClient(addr string, options ...func(c *Client) error) (*Client, error) {
	c := &Client{timeout: DefaultTimeout, network: "tcp", addr: addr}
	for _, f := range options {
		if f == nil {
			return nil, ErrNilOption
		}
		if err := f(c); err != nil {
			return nil, err
		}
	}
	if c.network == "tcp" {
		if !strings.Contains(c.addr, ":") {
			c.addr = fmt.Sprintf("%v:%v", c.addr, DefaultPort)
		}
	}
	var err error
	if c.conn, err = net.DialTimeout(c.network, c.addr, c.timeout); err != nil {
		return nil, err
	}

	c.scanner = bufio.NewScanner(bufio.NewReader(c.conn))
	c.scanner.Split(bufio.ScanLines)

	return c, nil
}

// setDeadline updates the deadline on the connection based on the clients configured timeout.
func (c *Client) setDeadline() error {
	return c.conn.SetDeadline(time.Now().Add(c.timeout))
}

// Exec executes cmd on the server and returns the response.
func (c *Client) Exec(cmd string) ([]string, error) {
	return c.ExecCmd(NewCmd(cmd))
}

// ExecCmd executes cmd on the server and returns the response.
func (c *Client) ExecCmd(cmd *Cmd) ([]string, error) {
	if err := c.setDeadline(); err != nil {
		return nil, err
	}

	if _, err := c.conn.Write([]byte(cmd.String())); err != nil {
		return nil, err
	}

	if err := c.setDeadline(); err != nil {
		return nil, err
	}

	if !c.scanner.Scan() {
		return nil, c.scanErr()
	}

	l := c.scanner.Text()
	matches := respRe.FindStringSubmatch(l)
	if len(matches) != 3 {
		return nil, NewInvalidResponseError("bad header", l)
	}

	cnt, err := strconv.Atoi(matches[1])
	if err != nil {
		// This should be impossible given the regexp matched.
		return nil, NewInvalidResponseError("bad header count", l)
	}

	switch {
	case cnt < 0:
		// rrdcached reported an error.
		return nil, NewError(cnt, matches[2])
	case cnt == 0:
		// message is the line e.g. first.
		return []string{matches[2]}, nil
	}

	if err := c.setDeadline(); err != nil {
		return nil, err
	}
	lines := make([]string, 0, cnt)
	for len(lines) < cnt && c.scanner.Scan() {
		lines = append(lines, c.scanner.Text())
		if err := c.setDeadline(); err != nil {
			return nil, err
		}
	}

	if len(lines) != cnt {
		// Short response.
		return nil, c.scanErr()
	}

	return lines, nil
}

// Close closes the connection to the server.
func (c *Client) Close() error {
	errD := c.setDeadline()
	_, errW := c.conn.Write([]byte("quit"))
	err := c.conn.Close()
	if err != nil {
		return err
	} else if errD != nil {
		return errD
	}

	return errW
}

// scanError returns the error from the scanner if non-nil,
// io.ErrUnexpectedEOF otherwise.
func (c *Client) scanErr() error {
	if err := c.scanner.Err(); err != nil {
		return err
	}
	return io.ErrUnexpectedEOF
}
