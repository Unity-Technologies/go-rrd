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
