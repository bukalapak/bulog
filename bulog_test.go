package bulog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
		{`level=INFO msg=info`, `level=INFO msg=info`},
		{`{"level":"INFO","msg":"info"}`, `{"level":"INFO","msg":"info"}`},
	},
	"NormalizeLevel": {
		{`[warn] warning`, `[WARN] warning`},
		{`level=WARN msg=warning`, `level=WARN msg=warning`},
		{`{"level":"WARN","msg":"warning"}`, `{"level":"WARN","msg":"warning"}`},
	},
	"SkipLevel": {
		{`[INFO] info`, `[DEBUG] debug`},
		{`level=INFO msg=info`},
		{`{"level":"INFO","msg":"info"}`},
	},
	"WithMetadata": {
		{`[INFO] info foo="bar" num=8 float=9.99 bool=true`},
		{`level=INFO bool=true float=9.99 foo=bar msg=info num=8`},
		{`{"level":"INFO","msg":"info","foo":"bar","num":8,"float":9.99,"bool":true}`},
	},
	"MsgMetadata": {
		{`[INFO] foo="bar" num=8 msg="info"`},
		{`level=INFO foo=bar msg=info num=8`},
		{`{"level":"INFO","msg":"info","foo":"bar","num":8}`},
	},
	"SpaceMetadata": {
		{`[INFO] foo="bar baz" info`},
		{`level=INFO foo="bar baz" msg=info`},
		{`{"level":"INFO","msg":"info","foo":"bar baz"}`},
	},
}

func TestOutput_Logfmt(t *testing.T) {
	doTest(t, bulog.Logfmt, 1)
}

func TestOutput_JSON(t *testing.T) {
	doTest(t, bulog.JSON, 2)
}

func doTest(t *testing.T, format bulog.Format, index int) {
	for k, v := range data {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = format
			l := log.New(w, "", 0)

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

	l := log.New(w, "", 0)
	l.Println("foo")

	c := struct {
		Timestamp time.Time `json:"@timestamp"`
	}{}

	b := w.Writer.(*bytes.Buffer).Bytes()

	json.Unmarshal(b, &c)

	if c.Timestamp.IsZero() {
		t.Fatal("bad time format")
	}
}

func TestOutput_caller(t *testing.T) {
	w := newOutput()
	w.Format = bulog.JSON
	w.ShowCaller = true

	l := log.New(w, "", 0)
	l.Println("foo")

	c := struct {
		Caller string `json:"caller"`
	}{}

	b := w.Writer.(*bytes.Buffer).Bytes()

	json.Unmarshal(b, &c)

	s := strings.Split(c.Caller, ":")

	if f := filepath.Base(s[0]); f != "bulog_test.go" {
		t.Fatal("bad caller")
	}
}

func TestOutput_stacktrace(t *testing.T) {
	w := newOutput()
	w.Format = bulog.JSON
	w.Stacktrace = true

	l := log.New(w, "", 0)
	l.Println("foo")

	c := struct {
		Stacktrace string `json:"stacktrace"`
	}{}

	b := w.Writer.(*bytes.Buffer).Bytes()

	json.Unmarshal(b, &c)

	if !strings.Contains(c.Stacktrace, "bulog_test.go") {
		t.Fatal("bad stacktrace")
	}
}

func TestOutput_log(t *testing.T) {
	data := []int{
		log.Ldate,
		log.Ltime,
		log.Lmicroseconds,
		log.Lshortfile,
		log.Llongfile,
		log.LstdFlags,
		log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile,
	}

	for _, flags := range data {
		t.Run(fmt.Sprintf("LOG: %d", flags), func(t *testing.T) {
			l := log.New(os.Stderr, "LOG: ", flags)
			w := newOutput()
			w.ShowCaller = true
			w.TimeFormat = time.RFC3339
			w.Format = bulog.JSON
			w.Attach(l)

			l.Println("[INFO] foo")

			c := struct {
				Msg       string    `json:"msg"`
				Level     string    `json:"level"`
				Caller    string    `json:"caller"`
				Timestamp time.Time `json:"@timestamp"`
			}{}

			b := w.Writer.(*bytes.Buffer).Bytes()
			json.Unmarshal(b, &c)

			if !strings.Contains(c.Caller, "bulog_test.go") {
				t.Fatal("bad caller")
			}

			if c.Level != "INFO" {
				t.Fatal("bad level")
			}

			if c.Timestamp.IsZero() {
				t.Fatal("bad time parsing")
			}

			if c.Msg != "LOG: foo" {
				t.Fatal("bad msg parsing")
			}
		})
	}
}

func newOutput() *bulog.Output {
	out := bulog.New("INFO", []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"})
	out.Writer = new(bytes.Buffer)
	out.ShowCaller = false
	out.Stacktrace = false
	out.TimeFormat = ""
	out.KeyNames = map[string]string{
		"timestamp": "@timestamp",
	}

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

func BenchmarkLogStd(b *testing.B) {
	l := log.New(ioutil.Discard, "INFO: ", log.LstdFlags)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		l.Println("hello")
	}
}

func BenchmarkLogfmt(b *testing.B) {
	out := bulog.New("INFO", []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"})
	out.Writer = ioutil.Discard

	l := log.New(out, "", 0)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		l.Println("hello")
	}
}

func BenchmarkJSON(b *testing.B) {
	out := bulog.New("INFO", []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"})
	out.Writer = ioutil.Discard
	out.Format = bulog.JSON

	l := log.New(out, "", 0)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		l.Println("hello")
	}
}
