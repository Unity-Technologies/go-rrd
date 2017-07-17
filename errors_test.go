package rrd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect string
	}{
		{"error", NewError(-1, "file not found"), "file not found (-1)"},
		{"invalid-response", NewInvalidResponseError("bad line", "bad line1", "bad line2"), "bad line (bad line1, bad line2)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.err.Error())
		})
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name string
		err  error
		f    func(error) bool
	}{
		{"exists", NewError(-1, "RRD Error: creating '/test.rrd': File exists"), IsExist},
		{"not-exists", NewError(-1, "No such file: /test-missing.rrd"), IsNotExist},
		{"illegal-update", NewError(-1, "illegal attempt to update using time 1499968801.000000 when last update time is 1499968801.000000 (minimum one second step)"), IsIllegalUpdate},
	}

	err2 := NewError(-1, "Some other error")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !assert.Error(t, tc.err) {
				return
			}
			assert.True(t, tc.f(tc.err))

			if !assert.Error(t, err2) {
				return
			}
			assert.False(t, tc.f(err2))
		})
	}
}
