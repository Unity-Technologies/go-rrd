package rrd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRRA(t *testing.T) {
	rawRRA := "RRA:AVERAGE:0.5:60:864000"
	tests := []struct {
		name   string
		rra    RRA
		expect string
	}{
		{"raw", NewRRA(rawRRA), rawRRA},
		{"average", NewAverage(0.5, 60, 129600), "RRA:AVERAGE:0.5:60:129600"},
		{"min", NewMin(0.5, 60, 129600), "RRA:MIN:0.5:60:129600"},
		{"max", NewMax(0.5, 60, 129600), "RRA:MAX:0.5:60:129600"},
		{"last", NewLast(0.5, 60, 129600), "RRA:LAST:0.5:60:129600"},
		{"hwpredict", NewHWPredict(10, 0.5, 0.5, 50, 1), "RRA:HWPREDICT:10:0.5:0.5:50:1"},
		{"mhwpredict", NewMHWPredict(10, 0.5, 0.5, 50, 1), "RRA:MHWPREDICT:10:0.5:0.5:50:1"},
		{"seasonal", NewSeasonal(10, 0.5, 1, 0.3), "RRA:SEASONAL:10:0.5:1:smoothing-window=0.3"},
		{"devseasonal", NewDevSeasonal(10, 0.5, 1, 0.3), "RRA:DEVSEASONAL:10:0.5:1:smoothing-window=0.3"},
		{"devpredict", NewDevPredict(10, 1), "RRA:DEVPREDICT:10:1"},
		{"failures", NewFailures(10, 3, 20, 1), "RRA:FAILURES:10:3:20:1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, string(tc.rra))
		})
	}
}
