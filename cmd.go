package rrd

import (
	"fmt"
)

// Cmd represents a rrdcached command.
type Cmd struct {
	cmd  string
	args []interface{}
}

// NewCmd creates a new Cmd.
func NewCmd(cmd string) *Cmd {
	return &Cmd{cmd: cmd}
}

// WithArgs sets the command Args.
func (c *Cmd) WithArgs(args ...interface{}) *Cmd {
	c.args = args
	return c
}

func (c *Cmd) String() string {
	args := append([]interface{}{c.cmd}, c.args...)
	return fmt.Sprintln(args...)
}
