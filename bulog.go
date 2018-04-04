package bulog

import (
	"bytes"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/go-logfmt/logfmt"
)

type Format int8

const (
	Basic Format = iota
	Logfmt
	JSON
)

type Output struct {
	Levels     []string
	MinLevel   string
	Format     Format
	Writer     io.Writer
	skipLevels map[string]struct{}
	once       sync.Once
}

func (w *Output) Write(line []byte) (n int, err error) {
	w.once.Do(w.init)

	l, z := w.extractLine(line)

	if _, ok := w.skipLevels[l]; ok {
		return len(line), nil
	}

	return w.Writer.Write(w.formatLine(l, z))
}

func (w *Output) init() {
	levels := make(map[string]struct{})
	for _, l := range w.Levels {
		if strings.EqualFold(l, w.MinLevel) {
			break
		}

		levels[strings.ToUpper(l)] = struct{}{}
	}

	w.skipLevels = levels
}

func (w *Output) extractLine(line []byte) (string, []byte) {
	level := w.extractLevel(line)

	if level != "" {
		return level, line[(len(level) + 3):]
	}

	return w.MinLevel, line
}

func (w *Output) extractLevel(line []byte) (level string) {
	x := bytes.IndexByte(line, '[')
	if x >= 0 {
		y := bytes.IndexByte(line[x:], ']')

		if y >= 0 {
			level = string(bytes.ToUpper(line[x+1 : x+y]))
		}
	}

	return
}

func (w *Output) formatLine(level string, line []byte) []byte {
	switch w.Format {
	case Logfmt:
		return w.formatLineFmt(level, line)
	case JSON:
		return w.formatLineJSON(level, line)
	}

	return w.formatLineNone(level, line)
}

func (w *Output) formatLineNone(level string, line []byte) []byte {
	b := new(bytes.Buffer)
	c := logfmt.NewEncoder(b)
	m := w.parseLine(line)
	q := w.sort(m)

	var msg []byte

	for _, k := range q {
		if k == "msg" {
			msg = m[k]
			continue
		}

		c.EncodeKeyval([]byte(k), m[k])
	}

	c.EndRecord()

	x := bytes.TrimSpace(b.Bytes())
	z := []byte("[" + level + "] ")
	z = append(z, msg...)

	if len(x) != 0 {
		z = append(z, []byte(" ")...)
		z = append(z, x...)
	}

	z = append(z, []byte("\n")...)

	return z
}

func (w *Output) formatLineFmt(level string, line []byte) []byte {
	b := new(bytes.Buffer)
	m := w.parseLine(line)
	q := w.sort(m)

	c := logfmt.NewEncoder(b)
	c.EncodeKeyval("level", level)

	for _, k := range q {
		c.EncodeKeyval([]byte(k), m[k])
	}

	c.EndRecord()

	return b.Bytes()
}

func (w *Output) formatLineJSON(level string, line []byte) []byte {
	b := []byte("{}")
	b, _ = jsonparser.Set(b, w.quote([]byte(level)), "level")

	for k, v := range w.parseLine(line) {
		b, _ = jsonparser.Set(b, w.autoQuote(v), k)
	}

	b = append(b, []byte("\n")...)

	return b
}

func (w *Output) quote(b []byte) []byte {
	return []byte(strconv.Quote(string(b)))
}

func (w *Output) autoQuote(b []byte) []byte {
	if w.quotable(b) {
		return w.quote(b)
	}

	return b
}

func (w *Output) quotable(b []byte) bool {
	if _, err := jsonparser.ParseInt(b); err == nil {
		return false
	}

	if _, err := jsonparser.ParseFloat(b); err == nil {
		return false
	}

	if _, err := jsonparser.ParseBoolean(b); err == nil {
		return false
	}

	return true
}

func (w *Output) parseLine(line []byte) map[string][]byte {
	m := make(map[string][]byte)
	d := logfmt.NewDecoder(bytes.NewReader(line))

	var hasMsg bool
	var msg [][]byte

	for d.ScanRecord() {
		for d.ScanKeyval() {
			if d.Value() != nil {
				if bytes.Equal(d.Key(), []byte("msg")) {
					hasMsg = true
				}

				m[string(d.Key())] = d.Value()
			} else {
				msg = append(msg, d.Key())
			}
		}
	}

	if !hasMsg {
		m["msg"] = bytes.Join(msg, []byte(" "))
	}

	return m
}

func (w *Output) sort(m map[string][]byte) []string {
	q := make([]string, len(m))
	i := 0

	for k := range m {
		q[i] = k
		i++
	}

	sort.Strings(q)

	return q
}
