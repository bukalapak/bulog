package bulog

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"github.com/go-logfmt/logfmt"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimestampFieldName = "@timestamp"
}

var mapLevel = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

type standard struct {
	logger zerolog.Logger
}

func (d *standard) Write(b []byte) (c int, err error) {
	n := len(b)

	if n > 0 && b[n-1] == '\n' {
		b = b[0 : n-1] // trim stdlog CR
	}

	z := bytes.Index(b, []byte("] "))

	if z > 1 {
		v := bytes.ToLower(b[1:z])
		b = b[z+2:]

		d.logger.WithLevel(mapLevel[string(v)]).Msg(string(b))
	}

	d.logger.WithLevel(zerolog.NoLevel).Msg(string(b))
	return
}

func newLog(out io.Writer) zerolog.Logger {
	return zerolog.New(out).With().Timestamp().Logger()
}

func newStandard(out io.Writer) *standard {
	return &standard{
		logger: newLog(out),
	}
}

func Standard(out io.Writer) *log.Logger {
	w := newStandard(out)
	l := log.New(w, "", 0)

	return l
}

type logFmt struct {
	logger zerolog.Logger
}

func newLogFmt(out io.Writer) *logFmt {
	return &logFmt{
		logger: newLog(out),
	}
}

func (m *logFmt) Write(b []byte) (n int, err error) {
	dec := logfmt.NewDecoder(bytes.NewReader(b))

	type kv struct {
		k, v []byte
	}

	var kvs []kv

	for dec.ScanRecord() {
		for dec.ScanKeyval() {
			k := dec.Key()
			v := dec.Value()

			if k != nil {
				kvs = append(kvs, kv{k, v})
			}
		}
	}

	lvl := zerolog.NoLevel
	msg := ""

	// extract level
	for i, x := range kvs {
		if bytes.EqualFold(x.k, []byte("level")) {
			lvl = mapLevel[string(x.v)]
			kvs = append(kvs[:i], kvs[i+1:]...)
		}
	}

	// extract msg
	for i, x := range kvs {
		if bytes.EqualFold(x.k, []byte("msg")) {
			msg = string(x.v)
			kvs = append(kvs[:i], kvs[i+1:]...)
		}
	}

	event := m.logger.WithLevel(lvl)

	for _, x := range kvs {
		if json.Valid(x.v) {
			event = event.RawJSON(string(x.k), x.v)
		} else {
			event = event.Bytes(string(x.k), x.v)
		}
	}

	event.Msg(msg)

	return
}

func LogFmt(out io.Writer) *log.Logger {
	w := newLogFmt(out)
	l := log.New(w, "", 0)

	return l
}
