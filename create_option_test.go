package rrd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateOption(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		opt    CreateOption
		expect string
	}{
		{"step", Step(time.Minute * 5), "-s 300"},
		{"start", Start(now), fmt.Sprintf("-b %v", now.Unix())},
		{"source", Source("test.rrd"), "-r test.rrd"},
		{"template", Template("test.rrd"), "-t test.rrd"},
		{"overwrite", Overwrite(), "-O"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, string(tc.opt))
		})
	}
}
