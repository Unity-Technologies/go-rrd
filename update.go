package rrd

import (
	"fmt"
	"strings"
	"time"
)

// Update represents a RRD update value.
type Update string

// NewUpdateRaw returns a new Update for the raw val.
// It supports the N: prefix for now times.
func NewUpdateRaw(val string) Update {
	if strings.HasPrefix(val, "N:") {
		val = fmt.Sprintf("%v:%v", time.Now().Unix(), val[2:])
	}
	return Update(val)
}

// NewUpdateNow returns a new update with the current time for the provided values.
func NewUpdateNow(val interface{}, values ...interface{}) Update {
	return NewUpdate(time.Now(), val, values...)
}

// NewUpdate returns a new update with the given ts for the provider values
func NewUpdate(ts time.Time, val interface{}, values ...interface{}) Update {
	parts := make([]string, len(values)+1)
	parts[0] = fmt.Sprint(val)
	for i, v := range values {
		parts[i+1] = fmt.Sprint(v)
	}
	return NewUpdateRaw(fmt.Sprintf("%v:%v", ts.Unix(), strings.Join(parts, ":")))
}
