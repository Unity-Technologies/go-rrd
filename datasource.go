package rrd

import (
	"fmt"
	"strings"
)

// Data Source Types
const (
	Gauge    = "GAUGE"
	Counter  = "COUNTER"
	DCounter = "DCOUNTER"
	Derive   = "DERIVE"
	DDerive  = "DDERIVE"
	Absolute = "ABSOLUTE"
	Compute  = "COMPUTE"
)

// Mapping applies a custom mapping to the data source.
func Mapping(name string, idx int) func(d *ds) {
	return func(d *ds) {
		d.mappedName = name
		d.sourceIndex = idx
	}
}

// DS represents a RRD data source.
type DS string

// NewDS returns a new DS for the value raw.
func NewDS(raw string) DS {
	return DS(raw)
}

// ds represents the internal data source that supports applying mappings
type ds struct {
	name        string
	mappedName  string
	sourceIndex int
	dst         string
	vals        []interface{}
}

func (d *ds) String() string {
	name := d.name
	if d.mappedName != "" {
		name += "=" + d.mappedName
		if d.sourceIndex != 0 {
			name += fmt.Sprintf("[%v]", d.sourceIndex)
		}
	}

	parts := make([]string, len(d.vals)+3)
	parts[0] = "DS"
	parts[1] = name
	parts[2] = d.dst
	for i, v := range d.vals {
		parts[i+3] = fmt.Sprint(v)
	}
	return strings.Join(parts, ":")
}

func newDS(dst, name string, options []func(d *ds), vals ...interface{}) DS {
	d := &ds{
		dst:  dst,
		name: name,
		vals: vals,
	}
	for _, f := range options {
		if f != nil {
			f(d)
		}
	}

	return DS(d.String())
}

// NewGauge returns a new GAUGE DS.
func NewGauge(name string, heartbeat, min, max int, options ...func(d *ds)) DS {
	return newDS(Gauge, name, options, heartbeat, min, max)
}

// NewCounter returns a new COUNTER DS.
func NewCounter(name string, heartbeat, min, max int, options ...func(d *ds)) DS {
	return newDS(Counter, name, options, heartbeat, min, max)
}

// NewDCounter returns a new DCOUNTER DS.
func NewDCounter(name string, heartbeat, min, max int, options ...func(d *ds)) DS {
	return newDS(DCounter, name, options, heartbeat, min, max)
}

// NewDerive returns a new DERIVE DS.
func NewDerive(name string, heartbeat, min, max int, options ...func(d *ds)) DS {
	return newDS(Derive, name, options, heartbeat, min, max)
}

// NewDDerive returns a new DDERIVE DS.
func NewDDerive(name string, heartbeat, min, max int, options ...func(d *ds)) DS {
	return newDS(DDerive, name, options, heartbeat, min, max)
}

// NewAbsolute returns a new ABSOLUTE DS.
func NewAbsolute(name string, heartbeat, min, max int, options ...func(d *ds)) DS {
	return newDS(Absolute, name, options, heartbeat, min, max)
}

// NewCompute returns a new COMPUTE DS.
func NewCompute(name, cdef string, options ...func(d *ds)) DS {
	return newDS(Compute, name, options, cdef)
}
