package rrd

import (
	"fmt"
	"time"
)

// CreateOption represents a rrd create option.
type CreateOption string

// Step returns a new create step option.
func Step(step time.Duration) CreateOption {
	return CreateOption(fmt.Sprintf("-s %v", int64(step/time.Second)))
}

// Start returns a new create start time option.
func Start(start time.Time) CreateOption {
	return CreateOption(fmt.Sprintf("-b %v", start.Unix()))
}

// Overwrite returns a new create overwrite option.
func Overwrite() CreateOption {
	return CreateOption("-O")
}

// Source returns a new create source option.
func Source(file string) CreateOption {
	return CreateOption(fmt.Sprintf("-r %v", file))
}

// Template returns a new create template option.
func Template(file string) CreateOption {
	return CreateOption(fmt.Sprintf("-t %v", file))
}
