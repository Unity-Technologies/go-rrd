package rrd

import (
	"fmt"
	"strings"
)

// Round Robin Algorithms
const (
	Average = "AVERAGE"
	Min     = "MIN"
	Max     = "MAX"
	Last    = "LAST"

	HoltWintersPredict          = "HWPREDICT"
	MultipliedHoltWinterPredict = "MHWPREDICT"
	Seasonal                    = "SEASONAL"
	DevSeasonal                 = "DEVSEASONAL"
	DevPredict                  = "DEVPREDICT"
	Failures                    = "FAILURES"
)

// RRA represents a raw RRA.
type RRA string

// NewRRA returns a new raw RRA set to val.
func NewRRA(val string) RRA {
	return RRA(val)
}

func newRRA(rra string, vals ...interface{}) RRA {
	parts := make([]string, len(vals)+2)
	parts[0] = "RRA"
	parts[1] = rra
	for i, v := range vals {
		parts[i+2] = fmt.Sprint(v)
	}
	return RRA(strings.Join(parts, ":"))
}

// NewAverage returns a new AVERAGE RRA.
func NewAverage(xff float32, steps, rows int) RRA {
	return newRRA(Average, xff, steps, rows)
}

// NewMin returns a new MIN RRA.
func NewMin(xff float32, steps, rows int) RRA {
	return newRRA(Min, xff, steps, rows)
}

// NewMax returns a new AVERAGE RRA.
func NewMax(xff float32, steps, rows int) RRA {
	return newRRA(Max, xff, steps, rows)
}

// NewLast returns a new AVERAGE RRA.
func NewLast(xff float32, steps, rows int) RRA {
	return newRRA(Last, xff, steps, rows)
}

// NewHWPredict returns a new HWPREDICT RRA.
func NewHWPredict(rows int, alpha, beta float32, period int, idx int) RRA {
	return newRRA(HoltWintersPredict, rows, alpha, beta, period, idx)
}

// NewMHWPredict returns a new HWPREDICT RRA.
func NewMHWPredict(rows int, alpha, beta float32, period int, idx int) RRA {
	return newRRA(MultipliedHoltWinterPredict, rows, alpha, beta, period, idx)
}

// NewSeasonal returns a new HWPREDICT RRA.
func NewSeasonal(period int, gamma float32, idx int, window float32) RRA {
	return newRRA(Seasonal, period, gamma, idx, fmt.Sprintf("smoothing-window=%v", window))
}

// NewDevSeasonal returns a new HWPREDICT RRA.
func NewDevSeasonal(period int, gamma float32, idx int, window float32) RRA {
	return newRRA(DevSeasonal, period, gamma, idx, fmt.Sprintf("smoothing-window=%v", window))
}

// NewDevPredict returns a new HWPREDICT RRA.
func NewDevPredict(rows, idx int) RRA {
	return newRRA(DevPredict, rows, idx)
}

// NewFailures returns a new HWPREDICT RRA.
func NewFailures(rows, threshold, window, idx int) RRA {
	return newRRA(Failures, rows, threshold, window, idx)
}
