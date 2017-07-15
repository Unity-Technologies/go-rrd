// +build integration

package rrd

import (
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	rrdAddress = flag.String("rrd-address", "127.0.0.1", "sets the rrdcached address for integration tests")
	rrdFile    = flag.String("rrd-file", "go-rrd-test.rrd", "sets the name of the RRD file for intergration tests")
)

func TestIntegration(t *testing.T) {
	c, err := NewClient(*rrdAddress, Timeout(time.Second*10))
	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		assert.NoError(t, c.Close())
	}()

	now := time.Now()

	invalid := func(t *testing.T) {
		l, err := c.Exec("invalid")
		assert.Error(t, err)
		assert.True(t, l == nil)
	}

	ping := func(t *testing.T) {
		assert.NoError(t, c.Ping())
	}

	create := func(t *testing.T) {
		err := c.Create(
			*rrdFile,
			[]DS{NewGauge("watts", time.Minute*5, 0, 1000)},
			[]RRA{NewAverage(0.5, 1, 864000)},
		)
		assert.NoError(t, err)
	}

	update := func(t *testing.T) {
		err := c.Update(*rrdFile, NewUpdate(now, 10))
		assert.NoError(t, err)
	}

	first := func(t *testing.T) {
		ts, err := c.First(*rrdFile, 0)
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, ts.Before(now))
	}

	last := func(t *testing.T) {
		ts, err := c.Last(*rrdFile)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, now.Truncate(time.Minute*5).Unix(), ts.Unix())
	}

	forget := func(t *testing.T) {
		assert.NoError(t, c.Forget(*rrdFile))
	}

	info := func(t *testing.T) {
		i, err := c.Info(*rrdFile)
		if !assert.NoError(t, err) {
			return
		}

		expected := map[string]interface{}{
			"step":                        int64(300),
			"ds[watts].index":             int64(0),
			"ds[watts].type":              "GAUGE",
			"ds[watts].minimal_heartbeat": int64(300),
			"ds[watts].min":               float64(0),
			"ds[watts].max":               float64(1000),
			"rra[0].cf":                   "AVERAGE",
			"rra[0].rows":                 int64(864000),
			"rra[0].xff":                  0.5,
		}

		for _, v := range i {
			if e, ok := expected[v.Key]; ok {
				assert.Equal(t, e, v.Value)
			} else if v.Key == "filename" {
				s := fmt.Sprint(v.Value)
				assert.True(t, strings.HasSuffix(s, *rrdFile))
			}

		}
	}

	flush := func(t *testing.T) {
		assert.NoError(t, c.Flush(*rrdFile))
	}

	fetch := func(t *testing.T) {
		f, err := c.Fetch(*rrdFile, Average)
		if !assert.NoError(t, err) {
			return
		}
		if assert.Len(t, f.Names, 1) {
			assert.Equal(t, "watts", f.Names[0])
		}
		assert.Equal(t, 1, f.Count)
		assert.Equal(t, time.Minute*5, f.Step)
	}

	fetchbin := func(t *testing.T) {
		f, err := c.FetchBin(*rrdFile, Average)
		if !assert.NoError(t, err) {
			return
		}
		if assert.Len(t, f.DS, 1) {
			assert.Equal(t, "watts", f.DS[0].Name)
		}
		assert.Equal(t, 1, f.Count)
		assert.Equal(t, time.Minute*5, f.Step)
	}

	stats := func(t *testing.T) {
		s, err := c.Stats()
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, s.UpdatesReceived > 0)
		assert.True(t, s.FlushesReceived > 0)
	}

	tests := []struct {
		name string
		f    func(t *testing.T)
	}{
		{"invalid", invalid},
		{"ping", ping},
		{"create", create},
		{"update", update},
		{"first", first},
		{"last", last},
		{"forget", forget},
		{"info", info},
		{"flush", flush},
		{"fetch", fetch},
		{"fetchbin", fetchbin},
		{"stats", stats},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.f)
	}
}
