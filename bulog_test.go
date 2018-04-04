package bulog_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bukalapak/bulog"
)

var data = map[string][][]string{
	"AutoLevel": {
		{`info`, `[INFO] info`},
		{`[INFO] info`, `[INFO] info`},
		{`level=INFO msg=info`, `level=INFO msg=info`},
		{`{"level":"INFO","msg":"info"}`, `{"level":"INFO","msg":"info"}`},
	},
	"NormalizeLevel": {
		{`[warn] warning`, `[WARN] warning`},
		{`[WARN] warning`, `[WARN] warning`},
		{`level=WARN msg=warning`, `level=WARN msg=warning`},
		{`{"level":"WARN","msg":"warning"}`, `{"level":"WARN","msg":"warning"}`},
	},
	"SkipLevel": {
		{`[INFO] info`, `[DEBUG] debug`},
		{`[INFO] info`},
		{`level=INFO msg=info`},
		{`{"level":"INFO","msg":"info"}`},
	},
	"WithMetadata": {
		{`[INFO] info foo="bar" num=8 float=9.99 bool=true`},
		{`[INFO] info bool=true float=9.99 foo=bar num=8`},
		{`level=INFO bool=true float=9.99 foo=bar msg=info num=8`},
		{`{"level":"INFO","msg":"info","foo":"bar","num":8,"float":9.99,"bool":true}`},
	},
	"MsgMetadata": {
		{`[INFO] foo="bar" num=8 msg="info"`},
		{`[INFO] info foo=bar num=8`},
		{`level=INFO foo=bar msg=info num=8`},
		{`{"level":"INFO","msg":"info","foo":"bar","num":8}`},
	},
	"SpaceMetadata": {
		{`[INFO] foo="bar baz" info`},
		{`[INFO] info foo="bar baz"`},
		{`level=INFO foo="bar baz" msg=info`},
		{`{"level":"INFO","msg":"info","foo":"bar baz"}`},
	},
}

func TestOutput(t *testing.T) {
	doTest(t, bulog.Basic, 1)
}

func TestOutput_Logfmt(t *testing.T) {
	doTest(t, bulog.Logfmt, 2)
}

func TestOutput_JSON(t *testing.T) {
	doTest(t, bulog.JSON, 3)
}

func doTest(t *testing.T, format bulog.Format, index int) {
	for k, v := range data {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = format
			l := bulog.NewLog(w)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[index], "\n") + "\n"

			if format == bulog.JSON {
				ss := strings.Split(strings.TrimSpace(s), "\n")
				xx := strings.Split(strings.TrimSpace(x), "\n")

				for i := range ss {
					if ok, _ := jsonEqual(ss[i], xx[i]); !ok {
						t.Fatalf("\nactual:   %q\nexpected: %q", s, x)
					}
				}
			} else {
				if s != x {
					t.Fatalf("\nactual:   %q\nexpected: %q", s, x)
				}
			}
		})
	}
}

func TestOutput_timestamp(t *testing.T) {
	w := newOutput()
	w.Format = bulog.JSON
	w.TimeFormat = time.RFC3339

	l := bulog.NewLog(w)
	l.Println("foo")

	c := struct {
		Timestamp time.Time `json:"timestamp"`
	}{}

	b := w.Writer.(*bytes.Buffer).Bytes()

	json.Unmarshal(b, &c)

	if c.Timestamp.IsZero() {
		t.Fatal("bad time format")
	}
}

func TestOutput_stacktrace(t *testing.T) {
	w := newOutput()
	w.Format = bulog.JSON
	w.Stacktrace = true

	l := bulog.NewLog(w)
	l.Println("foo")

	c := struct {
		Stacktrace string `json:"stacktrace"`
	}{}

	b := w.Writer.(*bytes.Buffer).Bytes()

	json.Unmarshal(b, &c)

	s := strings.Split(c.Stacktrace, ":")

	if f := filepath.Base(s[0]); f != "bulog_test.go" {
		t.Fatal("bad stacktrace")
	}
}

func newOutput() *bulog.Output {
	out := bulog.New("INFO", []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"})
	out.Writer = new(bytes.Buffer)
	out.Stacktrace = false
	out.TimeFormat = ""

	return out
}

func jsonEqual(s, x string) (bool, error) {
	var is interface{}
	var ix interface{}

	if err := json.Unmarshal([]byte(s), &is); err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(x), &ix); err != nil {
		return false, err
	}

	return reflect.DeepEqual(is, ix), nil
}
