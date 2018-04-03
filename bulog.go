package bulog

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/go-logfmt/logfmt"
)

type Format int8

const (
	None Format = iota
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
	d := logfmt.NewDecoder(bytes.NewReader(line))
	c := logfmt.NewEncoder(b)

	var hasMsg bool
	var msg [][]byte

	for d.ScanRecord() {
		for d.ScanKeyval() {
			if d.Value() != nil {
				if bytes.Equal(d.Key(), []byte("msg")) {
					hasMsg = true
				}

				c.EncodeKeyval(d.Key(), d.Value())
			} else {
				msg = append(msg, d.Key())
			}
		}
	}

	c.EndRecord()

	z := []byte("[" + level + "] ")
	x := bytes.TrimSpace(b.Bytes())

	if !hasMsg {
		z = append(z, bytes.Join(msg, []byte(" "))...)
	}

	if len(x) != 0 {
		z = append(z, []byte(" ")...)
		z = append(z, x...)
	}

	z = append(z, []byte("\n")...)

	return z
}

func (w *Output) formatLineFmt(level string, line []byte) []byte {
	b := new(bytes.Buffer)
	d := logfmt.NewDecoder(bytes.NewReader(line))
	c := logfmt.NewEncoder(b)
	c.EncodeKeyval("level", level)

	var hasMsg bool
	var msg [][]byte

	for d.ScanRecord() {
		for d.ScanKeyval() {
			if d.Value() != nil {
				if bytes.Equal(d.Key(), []byte("msg")) {
					hasMsg = true
				}

				c.EncodeKeyval(d.Key(), d.Value())
			} else {
				msg = append(msg, d.Key())
			}
		}
	}

	if !hasMsg {
		c.EncodeKeyval("msg", bytes.Join(msg, []byte(" ")))
	}

	c.EndRecord()

	return b.Bytes()
}

func (w *Output) formatLineJSON(level string, line []byte) []byte {
	d := logfmt.NewDecoder(bytes.NewReader(line))
	b := []byte("{}")
	b, _ = jsonparser.Set(b, w.quote([]byte(level)), "level")

	var hasMsg bool
	var msg [][]byte

	for d.ScanRecord() {
		for d.ScanKeyval() {
			if d.Value() != nil {
				if bytes.Equal(d.Key(), []byte("msg")) {
					hasMsg = true
				}

				b, _ = jsonparser.Set(b, w.autoQuote(d.Value()), string(d.Key()))
			} else {
				msg = append(msg, d.Key())
			}
		}
	}

	if !hasMsg {
		b, _ = jsonparser.Set(b, w.quote(bytes.Join(msg, []byte(" "))), "msg")
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
