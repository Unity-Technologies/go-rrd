package rrd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmd(t *testing.T) {
	tests := []struct {
		name   string
		cmd    *Cmd
		expect string
	}{
		{"ping", NewCmd("ping"), "ping"},
		{"flush", NewCmd("flush").WithArgs("test.rrd"), "flush test.rrd"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect+"\n", tc.cmd.String())
		})
	}
}
