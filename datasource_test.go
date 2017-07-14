package rrd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testDS = "a"

func TestDS(t *testing.T) {
	rawDS := "DS:a:GAUGE:120:0:100"
	tests := []struct {
		name   string
		ds     DS
		expect string
	}{
		{"gauge", NewGauge(testDS, 120, 0, 100), "DS:a:GAUGE:120:0:100"},
		{"counter", NewCounter(testDS, 120, 0, 100), "DS:a:COUNTER:120:0:100"},
		{"dcounter", NewDCounter(testDS, 120, 0, 100), "DS:a:DCOUNTER:120:0:100"},
		{"derive", NewDerive(testDS, 120, 0, 100), "DS:a:DERIVE:120:0:100"},
		{"dderive", NewDDerive(testDS, 120, 0, 100), "DS:a:DDERIVE:120:0:100"},
		{"absolute", NewAbsolute(testDS, 120, 0, 100), "DS:a:ABSOLUTE:120:0:100"},
		{"compute", NewCompute(testDS, "result=value,UN,0,value,IF"), "DS:a:COMPUTE:result=value,UN,0,value,IF"},
		{"mapped", NewGauge(testDS, 120, 0, 100, Mapping("b", 1)), "DS:a=b[1]:GAUGE:120:0:100"},
		{"raw", NewDS(rawDS), rawDS},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, string(tc.ds))
		})
	}
}
