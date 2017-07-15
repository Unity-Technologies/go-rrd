package rrd

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCmds(t *testing.T) {
	s := newServer(t)
	if s == nil {
		return
	}
	defer func() {
		assert.NoError(t, s.Close())
	}()

	c, err := NewClient(s.Addr, Timeout(time.Second*2))
	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		assert.NoError(t, c.Close())
	}()

	invalid := func(t *testing.T) {
		l, err := c.Exec("invalid")
		assert.Error(t, err)
		assert.True(t, l == nil)
	}

	ping := func(t *testing.T) {
		assert.NoError(t, c.Ping())
	}

	flush := func(t *testing.T) {
		assert.NoError(t, c.Flush("test.rrd"))
	}

	flushall := func(t *testing.T) {
		assert.NoError(t, c.FlushAll())
	}

	pending := func(t *testing.T) {
		_, err := c.Pending("test.rrd")
		assert.Error(t, err)
	}

	fetch := func(t *testing.T) {
		f, err := c.Fetch("test.rrd", Average)
		if !assert.NoError(t, err) {
			return
		}
		f1, f2 := float64(8), float64(1733.3512369791667)
		expected := &Fetch{
			FetchCommon: FetchCommon{
				FlushVersion: 1,
				Start:        time.Unix(1499908800, 0),
				End:          time.Unix(1499995500, 0),
				Step:         time.Minute * 5,
				Count:        2,
			},
			Names: []string{"watts", "amps"},
			Rows: []FetchRow{
				{
					Time: time.Unix(1499909100, 0),
					Data: []*float64{&f1, &f2},
				},
				{
					Time: time.Unix(1499909400, 0),
					Data: []*float64{nil, nil},
				},
			},
		}

		assert.Equal(t, expected, f)
	}

	fetchbin := func(t *testing.T) {
		f, err := c.FetchBin("test.rrd", Average)
		if !assert.NoError(t, err) {
			return
		}
		expected := &FetchBin{
			FetchCommon: FetchCommon{
				FlushVersion: 1,
				Start:        time.Unix(1499908800, 0),
				End:          time.Unix(1499995500, 0),
				Step:         time.Minute * 5,
				Count:        2,
			},
			DS: []*FetchBinDS{
				{
					Name:    "watts",
					Records: 1,
					Size:    8,
					Endian:  binary.LittleEndian,
					Data:    []interface{}{float64(0)},
				},
				{
					Name:    "amps",
					Records: 1,
					Size:    8,
					Endian:  binary.LittleEndian,
					Data:    []interface{}{float64(0)},
				},
			},
		}

		assert.Equal(t, expected, f)
	}

	forget := func(t *testing.T) {
		assert.NoError(t, c.Forget("test.rrd"))
	}

	queue := func(t *testing.T) {
		q, err := c.Queue("test.rrd")
		if !assert.NoError(t, err) {
			return
		}
		expected := []*Queue{
			{Size: 10, File: "test.rrd"},
		}
		assert.Equal(t, expected, q)
	}

	help := func(t *testing.T) {
		h, err := c.Help()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, commands["help"][1:], h)
	}

	stats := func(t *testing.T) {
		s, err := c.Stats()
		if !assert.NoError(t, err) {
			return
		}
		expected := &Stats{
			QueueLength:     0,
			UpdatesReceived: 1061847698,
			FlushesReceived: 1690,
			UpdatesWritten:  149201370,
			DataSetsWritten: 1061709625,
			TreeNodesNumber: 30727,
			TreeDepth:       18,
			JournalBytes:    0,
			JournalRotate:   0,
		}
		assert.Equal(t, expected, s)
	}

	update := func(t *testing.T) {
		assert.NoError(t, c.Update("test.rrd", "1499968801:U"))
	}

	wrote := func(t *testing.T) {
		assert.Error(t, c.Wrote("test.rrd"))
	}

	first := func(t *testing.T) {
		ts, err := c.First("test.rrd", 0)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, time.Unix(1240782000, 0), ts)
	}

	last := func(t *testing.T) {
		ts, err := c.Last("test.rrd")
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, time.Unix(1499981700, 0), ts)
	}

	info := func(t *testing.T) {
		i, err := c.Info("test.rrd")
		if !assert.NoError(t, err) {
			return
		}
		expected := []*Info{
			{Key: "filename", Value: "test.rrd"},
			{Key: "rrd_version", Value: "0003"},
			{Key: "step", Value: int64(300)},
			{Key: "last_update", Value: int64(1499981928)},
			{Key: "header_size", Value: int64(1760)},
			{Key: "ds[watts].index", Value: int64(0)},
			{Key: "ds[watts].type", Value: "GAUGE"},
			{Key: "ds[watts].minimal_heartbeat", Value: int64(300)},
			{Key: "ds[watts].min", Value: float64(0)},
			{Key: "ds[watts].max", Value: float64(24000)},
			{Key: "ds[watts].last_ds", Value: "U"},
			{Key: "ds[watts].unknown_sec", Value: int64(228)},
		}
		assert.Equal(t, expected, i)
	}

	create := func(t *testing.T) {
		err := c.Create(
			"test.rrd",
			[]DS{NewDS("DS:watts:GAUGE:300:0:24000")},
			[]RRA{NewRRA("RRA:AVERAGE:0.5:1:864000")},
		)
		assert.NoError(t, err)
	}

	batch := func(t *testing.T) {
		err := c.Batch(NewCmd("ping"), NewCmd("ping"))
		if !assert.Error(t, err) {
			return
		}
		assert.Equal(t, "1 Can't use 'ping' here.\n2 Can't use 'ping' here. (-2)", err.Error())
	}

	tests := []struct {
		name string
		f    func(t *testing.T)
	}{
		{"invalid", invalid},
		{"ping", ping},
		{"flush", flush},
		{"flushall", flushall},
		{"pending", pending},
		{"fetch", fetch},
		{"fetchbin", fetchbin},
		{"forget", forget},
		{"queue", queue},
		{"help", help},
		{"stats", stats},
		{"ping", ping},
		{"update", update},
		{"wrote", wrote},
		{"first", first},
		{"last", last},
		{"info", info},
		{"create", create},
		{"batch", batch},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.f)
	}
}
