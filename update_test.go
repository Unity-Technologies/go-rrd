package rrd

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	update := "1499995020:10:0.3"
	tests := []struct {
		name   string
		update Update
	}{
		{"raw", NewUpdateRaw(update)},
		{"raw-now", NewUpdateRaw("N:10:0.3")},
		{"basic", NewUpdate(time.Unix(1499995020, 0), 10, 0.3)},
		{"now", NewUpdateNow(10, 0.3)},
	}

	now := time.Now()
	eParts := strings.Split(update, ":")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Contains(tc.name, "now") {
				parts := strings.Split(string(tc.update), ":")
				if !assert.Equal(t, len(parts), len(eParts)) {
					return
				}
				for i, v := range parts {
					if i == 0 {
						i, err := strconv.ParseInt(v, 10, 64)
						if !assert.NoError(t, err) {
							return
						}
						assert.True(t, i >= now.Unix())
					} else {
						assert.Equal(t, eParts[i], v)
					}
				}
			} else {
				assert.Equal(t, update, string(tc.update))
			}
		})
	}
}
