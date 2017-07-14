package rrd

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	cmdQuit = "quit"
)

var (
	commands = map[string][]string{
		"ping":     {"0 PONG"},
		"flush":    {"0 Nothing to flush: /test.rrd."},
		"flushall": {"0 Started flush."},
		"pending":  {"-1 No such file or directory."},
		"fetch": {
			"8 Success",
			"FlushVersion: 1",
			"Start: 1499908800",
			"End: 1499995500",
			"Step: 300",
			"DSCount: 2",
			"DSName: watts amps",
			"1499909100: 8.00000000000000000e+00 1.73335123697916674e+03",
			"1499909400: nan nan",
		},
		"fetchbin": {
			"7 Success",
			"FlushVersion: 1",
			"Start: 1499908800",
			"End: 1499995500",
			"Step: 300",
			"DSCount: 2",
			"DSName-watts: BinaryData 1441 8 LITTLE",
			"DSName-amps: BinaryData 1441 8 LITTLE",
		},
		"forget": {"0 Gone!"},
		"queue":  {"1 in queue.", "10 test.rrd"},
		"help": {
			"4 Help for QUIT",
			"Usage: QUIT",
			"",
			"Disconnect from rrdcached.",
			"",
		},
		"stats": {
			"9 Statistics follow",
			"QueueLength: 0",
			"UpdatesReceived: 1061847698",
			"FlushesReceived: 1690",
			"UpdatesWritten: 149201370",
			"DataSetsWritten: 1061709625",
			"TreeNodesNumber: 30727",
			"TreeDepth: 18",
			"JournalBytes: 0",
			"JournalRotate: 0",
		},
		"update": {"0 errors, enqueued 1 value(s)."},
		"wrote":  {"-1 Can't use 'wrote' here."},
		"first":  {"0 1240782000"},
		"last":   {"0 1499981700"},
		"info": {
			"12 Info for test.rrd follows",
			"filename 2 test.rrd",
			"rrd_version 2 0003",
			"step 1 300",
			"last_update 1 1499981928",
			"header_size 1 1760",
			"ds[watts].index 1 0",
			"ds[watts].type 2 GAUGE",
			"ds[watts].minimal_heartbeat 1 300",
			"ds[watts].min 0 0.0000000000e+00",
			"ds[watts].max 0 2.4000000000e+04",
			"ds[watts].last_ds 2 U",
			"ds[watts].unknown_sec 1 228",
		},
		"create": {"0 RRD created OK"},
		"batch":  {"0 Go ahead.  End with dot '.' on its own line."},
		".": {
			"2 errors",
			"1 Can't use 'ping' here.",
			"2 Can't use 'ping' here.",
		},
		cmdQuit: nil,
	}
)

// newLockListener creates a new listener on the local IP.
func newLocalListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// server is a mock TeamSpeak 3 server
type server struct {
	Addr     string
	Listener net.Listener

	t        *testing.T
	conns    map[net.Conn]struct{}
	done     chan struct{}
	wg       sync.WaitGroup
	failConn bool
	mtx      sync.Mutex
}

// sconn represents a server connection
type sconn struct {
	id int
	net.Conn
}

// newServer returns a running server or nil if an error occurred.
func newServer(t *testing.T) *server {
	s := newServerStopped(t)
	s.Start()

	return s
}

// newServerStopped returns a stopped servers or nil if an error occurred.
func newServerStopped(t *testing.T) *server {
	l, err := newLocalListener()
	if !assert.NoError(t, err) {
		return nil
	}

	s := &server{
		Listener: l,
		conns:    make(map[net.Conn]struct{}),
		done:     make(chan struct{}),
		t:        t,
	}
	s.Addr = s.Listener.Addr().String()
	return s
}

// Start starts the server.
func (s *server) Start() {
	s.wg.Add(1)
	go s.serve()
}

// server processes incoming requests until signaled to stop with Close.
func (s *server) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.running() {
				assert.NoError(s.t, err)
			}
			return
		}
		s.wg.Add(1)
		go s.handle(conn)
	}
}

// write writes msg to conn.
func (s *server) write(conn net.Conn, lines ...string) error {
	_, err := conn.Write([]byte(strings.Join(lines, "\n") + "\n"))
	if s.running() {
		assert.NoError(s.t, err)
	}

	return err
}

// running returns true unless Close has been called, false otherwise.
func (s *server) running() bool {
	select {
	case <-s.done:
		return false
	default:
		return true
	}
}

// handle handles a client connection.
func (s *server) handle(conn net.Conn) {
	s.mtx.Lock()
	s.conns[conn] = struct{}{}
	s.mtx.Unlock()
	defer func() {
		s.closeConn(conn)
		s.wg.Done()
	}()

	if s.failConn {
		return
	}

	sc := bufio.NewScanner(bufio.NewReader(conn))
	sc.Split(bufio.ScanLines)

	c := &sconn{Conn: conn}
	var batch bool
	for sc.Scan() {
		l := sc.Text()
		parts := strings.Split(l, " ")
		switch parts[0] {
		case cmdQuit:
			return
		case ".":
			batch = false
		}

		if batch {
			continue
		}

		resp, ok := commands[parts[0]]
		var err error
		if ok {
			err = s.write(c, resp...)
		} else {
			err = s.write(c, fmt.Sprintf("-1 Unknown command: %v", parts[0]))
		}
		if err != nil {
			return
		}
		if parts[0] == "batch" {
			batch = true
		}
	}

	if err := sc.Err(); err != nil && s.running() {
		assert.NoError(s.t, err)
	}
}

// closeConn closes a client connection and removes it from our map of connections.
func (s *server) closeConn(conn net.Conn) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	conn.Close() // nolint: errcheck
	delete(s.conns, conn)
}

// Close cleanly shuts down the server.
func (s *server) Close() error {
	close(s.done)
	err := s.Listener.Close()
	s.mtx.Lock()
	for c := range s.conns {
		if err2 := c.Close(); err2 != nil && err == nil {
			err = err2
		}
	}
	s.mtx.Unlock()
	s.wg.Wait()

	return err
}
