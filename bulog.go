package bulog

import (
	"bytes"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-logfmt/logfmt"
)

type Format int8

const (
	Logfmt Format = iota
	JSON
)

type Output struct {
	Levels     []string
	MinLevel   string
	Format     Format
	Writer     io.Writer
	TimeFormat string
	ShowCaller bool
	Stacktrace bool

	logPrefix  string
	logFlags   int
	skipLevels map[string]struct{}
	once       sync.Once
}

func New(minLevel string, levels []string) *Output {
	return &Output{
		Levels:     levels,
		MinLevel:   minLevel,
		Writer:     os.Stderr,
		Format:     Logfmt,
		TimeFormat: time.RFC3339,
		ShowCaller: true,
		Stacktrace: true,
	}
}

func NewLog(w io.Writer) *log.Logger {
	return log.New(w, "", 0)
}

func (w *Output) Attach(g *log.Logger) {
	w.logPrefix = g.Prefix()
	w.logFlags = g.Flags()
	g.SetOutput(w)
}

func (w *Output) Write(line []byte) (n int, err error) {
	w.once.Do(w.init)

	level := w.extractLevel(line)

	if _, ok := w.skipLevels[w.currentLevel(level)]; ok {
		return len(line), nil
	}

	return w.Writer.Write(w.formatLine(level, line))
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

func (w *Output) currentLevel(level string) string {
	if level == "" {
		return w.MinLevel
	}

	return level
}

func (w *Output) extractLevel(line []byte) string {
	x := bytes.IndexByte(line, '[')
	if x >= 0 {
		y := bytes.IndexByte(line[x:], ']')

		if y >= 0 {
			return string(bytes.ToUpper(line[x+1 : x+y]))
		}
	}

	return ""
}

func (w *Output) formatLine(level string, line []byte) []byte {
	switch w.Format {
	case JSON:
		return w.formatLineJSON(level, line)
	case Logfmt:
		fallthrough
	default:
		return w.formatLineLogfmt(level, line)
	}
}

func (w *Output) formatLineLogfmt(level string, line []byte) []byte {
	b := new(bytes.Buffer)
	m := w.parseLine(level, line)
	q := sortStrings(m)
	l := w.currentLevel(level)
	c := logfmt.NewEncoder(b)
	c.EncodeKeyval("level", l)

	for _, k := range q {
		c.EncodeKeyval([]byte(k), m[k])
	}

	c.EndRecord()

	return b.Bytes()
}

func (w *Output) formatLineJSON(level string, line []byte) []byte {
	b := []byte("{}")
	l := w.currentLevel(level)

	b, _ = jsonparser.Set(b, quote([]byte(l)), "level")

	for k, v := range w.parseLine(level, line) {
		b, _ = jsonparser.Set(b, quotable(v), k)
	}

	b = append(b, '\n')

	return b
}

func (w *Output) parseLine(level string, line []byte) map[string][]byte {
	line = w.stripPrefix(line)
	now, line := w.extractTimestamp(line)
	caller, line := w.extractCaller(line)

	d := logfmt.NewDecoder(bytes.NewReader(line))
	m := make(map[string][]byte)

	if w.TimeFormat != "" {
		m["timestamp"] = []byte(now.Format(w.TimeFormat))
	}

	if w.ShowCaller {
		m["caller"] = caller
	}

	if w.Stacktrace {
		m["stacktrace"] = debug.Stack()
	}

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
		msg := bytes.Join(msg, []byte(" "))
		if level != "" {
			m["msg"] = msg[len(level)+3:]
		} else {
			m["msg"] = msg
		}
	}

	if w.logPrefix != "" {
		m["msg"] = append([]byte(w.logPrefix), m["msg"]...)
	}

	return m
}

func (w *Output) stripPrefix(line []byte) []byte {
	if w.logPrefix != "" {
		return line[len(w.logPrefix):]
	}

	return line
}

func (w *Output) extractTimestamp(line []byte) (time.Time, []byte) {
	if w.logFlags&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		var layout string

		if w.logFlags&log.Ldate != 0 {
			layout += "2006/01/02"
		}

		if w.logFlags&(log.Ltime|log.Lmicroseconds) != 0 {
			if w.logFlags&log.Ldate != 0 {
				layout += " "
			}

			layout += "15:04:05"

			if w.logFlags&log.Lmicroseconds != 0 {
				layout += ".000000"
			}
		}

		n := len(layout)

		t, err := time.Parse(layout, string(line[:n]))
		if err == nil {
			return t, line[n+1:]
		}
	}

	return time.Now(), line
}

func (w *Output) extractCaller(line []byte) ([]byte, []byte) {
	if w.logFlags&(log.Lshortfile|log.Llongfile) != 0 {
		b := bytes.SplitN(line, []byte(" "), 2)
		return b[0][:len(b[0])-1], b[1]
	}

	return caller(), line
}

func sortStrings(m map[string][]byte) []string {
	q := make([]string, len(m))
	i := 0

	for k := range m {
		q[i] = k
		i++
	}

	sort.Strings(q)

	return q
}

func quote(b []byte) []byte {
	return []byte(strconv.Quote(string(b)))
}

func quotable(b []byte) []byte {
	if isQuotable(b) {
		return quote(b)
	}

	return b
}

func isQuotable(b []byte) bool {
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

func caller() []byte {
	_, file, line, _ := runtime.Caller(8)
	c := []byte(file)
	c = append(c, ':')
	c = append(c, []byte(strconv.Itoa(line))...)
	return c
}
