package rrd

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	valueRe = regexp.MustCompile(`^(\w+):\s+(\w+)`)
)

// Flush requests rrdcached flushed all values pending for filename to disk.
func (c *Client) Flush(filename string) error {
	_, err := c.ExecCmd(NewCmd("flush").WithArgs(filename))
	return err
}

// FlushAll requests the rrdcached start to flush all pending values to disk.
func (c *Client) FlushAll() error {
	_, err := c.Exec("flushall")
	return err
}

// Pending returns any "pending" updates for a file, in order.
func (c *Client) Pending(filename string) ([]string, error) {
	// TODO(steve): parse if needed when we know what the data looks like.
	return c.ExecCmd(NewCmd("pending").WithArgs(filename))
}

// FetchCommon represents the common fields between fetch and fetchbin
type FetchCommon struct {
	FlushVersion int
	Start        time.Time
	End          time.Time
	Step         time.Duration
	Count        int
}

// Fetch represents the data returned by an rrdcached fetch command.
type Fetch struct {
	FetchCommon
	Names []string
	Rows  []FetchRow
}

// FetchRow represents a single row of data returned by an rrdcached fetch command.
type FetchRow struct {
	Time time.Time
	Data []*float64
}

// decodeField decodes val into the field.
// It supports int, int64, string, time.Time, time.Duration fields only.
func decodeField(field, val, line string, fv reflect.Value) error {
	switch fv.Kind() {
	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, fv.Type().Bits())
		if err != nil {
			return NewInvalidResponseError(fmt.Sprintf("decodeField: invalid int %v for field %v", val, field), line)
		}
		if _, ok := fv.Interface().(time.Duration); ok {
			fv.SetInt(int64(time.Second) * i)
		} else {
			fv.SetInt(i)
		}
	case reflect.Float64:
		f, err := strconv.ParseFloat(val, fv.Type().Bits())
		if err != nil {
			return NewInvalidResponseError(fmt.Sprintf("decodeField: invalid float %v for field %v", val, field), line)
		}
		fv.SetFloat(f)
	case reflect.String:
		fv.SetString(val)
	case reflect.Struct:
		i := fv.Interface()
		switch i := i.(type) {
		case time.Time:
			if err := decodeTime(field, val, line, fv); err != nil {
				return err
			}
		default:
			return NewInvalidResponseError(fmt.Sprintf("decodeField: unsupported type %v for field %v", i, field), line)
		}
	default:
		return fmt.Errorf("unsupported type %T for field %v", fv, field)
	}

	return nil
}

// decodeTime decodes a unix timestamp into a time.Time field.
func decodeTime(field, val, line string, fv reflect.Value) error {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return NewInvalidResponseError(fmt.Sprintf("decodeTime: invalid int %v for field %v", val, field), line)
	}
	t := time.Unix(i, 0)
	fv.Set(reflect.ValueOf(t))

	return nil
}

// fetch performs the common action between fetch and fetchbin.
func (c *Client) fetch(cmd, filename, cf string, r interface{}, options ...interface{}) ([]string, error) {
	args := append([]interface{}{filename, cf}, options...)
	lines, err := c.ExecCmd(NewCmd(cmd).WithArgs(args...))
	if err != nil {
		return nil, err
	}

	v := reflect.Indirect(reflect.ValueOf(r))
	for i, l := range lines {
		if strings.HasPrefix(l, "DSName:") {
			r, ok := r.(*Fetch)
			if !ok {
				return nil, NewInvalidResponseError(cmd+": unexpected ds name", l)
			}
			l = l[7:]
			l = strings.TrimSpace(l)
			r.Names = strings.Split(l, " ")
			if len(r.Names) != r.Count {
				return nil, NewInvalidResponseError(cmd+": invalid ds name count", l)
			}
			return lines[i+1:], nil
		} else if strings.HasPrefix(l, "DSName-") {
			return lines[i:], nil
		} else if matches := valueRe.FindStringSubmatch(l); len(matches) == 3 {
			field := strings.TrimPrefix(matches[1], "DS")
			fv := v.FieldByName(field)
			if !fv.IsValid() {
				return nil, NewInvalidResponseError("unknown field", l)
			}
			if err := decodeField(field, matches[2], l, fv); err != nil {
				return nil, err
			}
		}
	}

	return nil, NewInvalidResponseError(cmd+": missing ds name", lines...)
}

// Fetch returns the free text results of a fetch command with the given options.
func (c *Client) Fetch(filename, cf string, options ...interface{}) (*Fetch, error) {
	r := &Fetch{}
	lines, err := c.fetch("fetch", filename, cf, r, options...)
	if err != nil {
		return nil, err
	}

	for _, l := range lines {
		parts := strings.SplitN(l, ":", 2)
		if len(parts) != 2 {
			return nil, NewInvalidResponseError("fetch: unsupported value", l)
		}

		i, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, NewInvalidResponseError("fetch: invalid ds", l)
		}

		fr := FetchRow{
			Time: time.Unix(i, 0),
			Data: make([]*float64, len(r.Names)),
		}
		for i, val := range strings.Split(strings.TrimSpace(parts[1]), " ") {
			if val == "nan" {
				continue
			}

			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, NewInvalidResponseError("fetch: invalid ds val", l)
			}
			fr.Data[i] = &v
		}
		r.Rows = append(r.Rows, fr)
	}

	return r, nil
}

// FetchBinDS represents a row of binary data.
type FetchBinDS struct {
	Name    string
	Records int
	Size    int
	Endian  binary.ByteOrder
}

// FetchBin represents the response from a fetchbin command.
type FetchBin struct {
	FetchCommon
	DS []*FetchBinDS
}

// newFetchBinDS returns a new FetchBinDS created from the data in line.
func newFetchBinDS(line string) (*FetchBinDS, error) {
	r := &FetchBinDS{}
	parts := strings.Split(line, " ")
	if len(parts) != 5 {
		return nil, NewInvalidResponseError("fetchbin: row invalid ds", line)
	}

	v, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, NewInvalidResponseError("fetchbin: row invalid records", line)
	}
	r.Records = v

	if v, err = strconv.Atoi(parts[3]); err != nil {
		return nil, NewInvalidResponseError("fetchbin: row invalid size", line)
	}
	r.Size = v
	r.Name = parts[0][7 : len(parts[0])-1]
	switch parts[4] {
	case "LITTLE":
		r.Endian = binary.LittleEndian
	case "BIG":
		r.Endian = binary.BigEndian
	default:
		return nil, NewInvalidResponseError("fetchbin: row invalid endian", line)
	}

	return r, nil
}

// FetchBin returns the text/binary results of a fetch command with the given options.
func (c *Client) FetchBin(filename, cf string, options ...interface{}) (*FetchBin, error) {
	r := &FetchBin{}
	lines, err := c.fetch("fetchbin", filename, cf, r, options...)
	if err != nil {
		return nil, err
	}

	if len(lines) != r.Count {
		return nil, NewInvalidResponseError("fetchbin: invalid ds count", lines...)
	}

	r.DS = make([]*FetchBinDS, len(lines))
	for i, l := range lines {
		var err error
		r.DS[i], err = newFetchBinDS(l)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Forget requests rrdcached remove filename from the cache.
// Any pending updates WILL BE LOST.
func (c *Client) Forget(filename string) error {
	_, err := c.ExecCmd(NewCmd("forget").WithArgs(filename))
	return err
}

// Queue represents a file queue.
type Queue struct {
	Size int64
	File string
}

// Queue returns the files that are on the rrdcached output queue.
func (c *Client) Queue(filename string) ([]*Queue, error) {
	lines, err := c.ExecCmd(NewCmd("queue").WithArgs(filename))
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, nil
	}

	queued := make([]*Queue, len(lines))
	for i, l := range lines {
		parts := strings.SplitN(l, " ", 2)
		if len(parts) != 2 {
			return nil, NewInvalidResponseError("queue: invalid parts", l)
		}
		v, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return nil, NewInvalidResponseError("queue: invalid num", l)
		}

		queued[i] = &Queue{Size: v, File: strings.TrimSpace(parts[1])}
	}

	return queued, nil
}

// Help returns command help.
func (c *Client) Help(cmd ...string) ([]string, error) {
	switch len(cmd) {
	case 0:
		return c.Exec("help")
	case 1:
		return c.ExecCmd(NewCmd("help").WithArgs(cmd[0]))
	default:
		return nil, fmt.Errorf("more than one cmd specified")
	}
}

// Stats represents rrdcached stats.
type Stats struct {
	QueueLength     int64
	UpdatesReceived int64
	FlushesReceived int64
	UpdatesWritten  int64
	DataSetsWritten int64
	TreeNodesNumber int64
	TreeDepth       int64
	JournalBytes    int64
	JournalRotate   int64
}

// Stats returns stats about rrdcached.
func (c *Client) Stats() (*Stats, error) {
	lines, err := c.Exec("stats")
	if err != nil {
		return nil, err
	}

	s := &Stats{}
	v := reflect.Indirect(reflect.ValueOf(s))
	for _, l := range lines {
		if matches := valueRe.FindStringSubmatch(l); len(matches) == 3 {
			i, err := strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				return nil, NewInvalidResponseError("stats: invalid val", l)
			}
			if f := v.FieldByName(matches[1]); f.IsValid() {
				f.SetInt(i)
			}
		} else {
			return nil, NewInvalidResponseError("stats: unrecognised line", l)
		}
	}

	return s, nil
}

// Ping sends a ping to the server.
func (c *Client) Ping() error {
	_, err := c.Exec("ping")
	return err
}

// Update adds more data to filename.
func (c *Client) Update(filename string, value Update, values ...Update) error {
	args := make([]interface{}, len(values)+2)
	args[0] = filename
	args[1] = value
	for i, v := range values {
		args[i+2] = v
	}
	_, err := c.ExecCmd(NewCmd("update").WithArgs(args...))
	return err
}

// Wrote sends a wrote command for filename to rrdcached.
func (c *Client) Wrote(filename string) error {
	_, err := c.ExecCmd(NewCmd("wrote").WithArgs(filename))
	return err
}

// First returns the timestamp of the first CDP for the given RRA.
func (c *Client) First(filename string, rra int) (time.Time, error) {
	return c.parseTime(c.ExecCmd(NewCmd("first").WithArgs(filename, rra)))
}

// partsTime parses the time stored in the first line and returns it
func (c *Client) parseTime(lines []string, err error) (time.Time, error) {
	var t time.Time
	if err != nil {
		return t, err
	}

	if len(lines) != 1 {
		return t, NewInvalidResponseError("parseTime: unexpected lines", lines...)
	}

	i, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		return t, NewInvalidResponseError("parseTime: parse int", lines[0])
	}

	return time.Unix(i, 0), nil
}

// Last returns the timestamp of the last update to the specified RRD.
func (c *Client) Last(filename string) (time.Time, error) {
	return c.parseTime(c.ExecCmd(NewCmd("last").WithArgs(filename)))
}

// Info represents the configuration information of an RRD.
type Info struct {
	Key   string
	Value interface{}
}

// Info returns the configuration information for the specified RRD.
func (c *Client) Info(filename string) ([]*Info, error) {
	lines, err := c.ExecCmd(NewCmd("info").WithArgs(filename))
	if err != nil {
		return nil, err
	}

	data := make([]*Info, len(lines))
	for i, l := range lines {
		parts := strings.SplitN(l, " ", 3)
		if len(parts) != 3 {
			return nil, NewInvalidResponseError("info: invalid parts", l)
		}
		info := &Info{Key: parts[0]}
		switch parts[1] {
		case "2":
			// string
			info.Value = parts[2]
		case "1":
			// int
			v, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, NewInvalidResponseError(fmt.Sprintf("info: invalid int for key %v", info.Key), l)
			}
			info.Value = v
		case "0":
			// float
			v, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return nil, NewInvalidResponseError(fmt.Sprintf("info: invalid float for key %v", info.Key), l)
			}
			info.Value = v
		default:
			return nil, NewInvalidResponseError(fmt.Sprintf("info: unknown type %v for key %v", parts[1], info.Key), l)
		}
		data[i] = info
	}

	return data, nil
}

// Create creates the RRD according to the supplied parameters.
func (c *Client) Create(filename string, ds []DS, rra []RRA, options ...interface{}) error {
	args := append([]interface{}{filename}, options...)
	for _, v := range ds {
		args = append(args, v)
	}
	for _, v := range rra {
		args = append(args, v)
	}
	_, err := c.ExecCmd(NewCmd("create").WithArgs(args...))
	return err
}

// Batch initiates the bulk load of multiple commands.
func (c *Client) Batch(cmds ...*Cmd) error {
	_, err := c.Exec("batch")
	if err != nil {
		return err
	}

	lines := make([]string, len(cmds)+1)
	for i, c := range cmds {
		lines[i] = c.String()
	}
	lines[len(cmds)] = ".\n"

	if err = c.setDeadline(); err != nil {
		return err
	}

	if _, err = c.conn.Write([]byte(strings.Join(lines, ""))); err != nil {
		return err
	}

	if err = c.setDeadline(); err != nil {
		return err
	}

	if !c.scanner.Scan() {
		return c.scanErr()
	}

	l := c.scanner.Text()
	matches := respRe.FindStringSubmatch(l)
	if len(matches) != 3 {
		return NewInvalidResponseError("batch: invalid matches", l)
	}

	cnt, err := strconv.Atoi(matches[1])
	if err != nil {
		// This should be impossible given the regexp matched.
		return NewInvalidResponseError("batch: invalid count", l)
	}

	if cnt == 0 {
		return nil
	}

	if err := c.setDeadline(); err != nil {
		return err
	}
	rlines := make([]string, 0, cnt)
	for c.scanner.Scan() && len(rlines) < cnt {
		rlines = append(rlines, c.scanner.Text())
		if err := c.setDeadline(); err != nil {
			return err
		}
	}

	if len(rlines) != cnt {
		// Short response.
		return c.scanErr()
	}

	return NewError(0-cnt, strings.Join(rlines, "\n"))
}
