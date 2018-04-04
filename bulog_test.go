package bulog_test

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/bukalapak/bulog"
)

var data = map[string][][]string{
	"AutoLevel": [][]string{
		[]string{`info`, `[INFO] info`},
		[]string{`[INFO] info`, `[INFO] info`},
		[]string{`level=INFO msg=info`, `level=INFO msg=info`},
		[]string{`{"level":"INFO","msg":"info"}`, `{"level":"INFO","msg":"info"}`},
	},
	"NormalizeLevel": [][]string{
		[]string{`[warn] warning`, `[WARN] warning`},
		[]string{`[WARN] warning`, `[WARN] warning`},
		[]string{`level=WARN msg=warning`, `level=WARN msg=warning`},
		[]string{`{"level":"WARN","msg":"warning"}`, `{"level":"WARN","msg":"warning"}`},
	},
	"SkipLevel": [][]string{
		[]string{`[INFO] info`, `[DEBUG] debug`},
		[]string{`[INFO] info`},
		[]string{`level=INFO msg=info`},
		[]string{`{"level":"INFO","msg":"info"}`},
	},
	"WithMetadata": [][]string{
		[]string{`[INFO] info foo="bar" num=8`},
		[]string{`[INFO] info foo=bar num=8`},
		[]string{`level=INFO foo=bar msg=info num=8`},
		[]string{`{"level":"INFO","msg":"info","foo":"bar","num":8}`},
	},
	"MsgMetadata": [][]string{
		[]string{`[INFO] foo="bar" num=8 msg="info"`},
		[]string{`[INFO] info foo=bar num=8`},
		[]string{`level=INFO foo=bar msg=info num=8`},
		[]string{`{"level":"INFO","msg":"info","foo":"bar","num":8}`},
	},
	"SpaceMetadata": [][]string{
		[]string{`[INFO] foo="bar baz" info`},
		[]string{`[INFO] info foo="bar baz"`},
		[]string{`level=INFO foo="bar baz" msg=info`},
		[]string{`{"level":"INFO","msg":"info","foo":"bar baz"}`},
	},
}

func TestOutput(t *testing.T) {
	for k, v := range data {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = bulog.Basic
			l := log.New(w, "", 0)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[1], "\n") + "\n"

			if s != x {
				t.Fatalf("\nactual:   %q\nexpected: %q", s, x)
			}
		})
	}
}

func TestOutput_Logfmt(t *testing.T) {
	for k, v := range data {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = bulog.Logfmt
			l := log.New(w, "", 0)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[2], "\n") + "\n"

			if s != x {
				t.Fatalf("\nactual:   %q\nexpected: %q", s, x)
			}
		})
	}
}

func TestOutput_JSON(t *testing.T) {
	for k, v := range data {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = bulog.JSON
			l := log.New(w, "", 0)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[3], "\n") + "\n"

			ss := strings.Split(strings.TrimSpace(s), "\n")
			xx := strings.Split(strings.TrimSpace(x), "\n")

			for i := range ss {
				if ok, _ := jsonEqual(ss[i], xx[i]); !ok {
					t.Fatalf("\nactual:   %q\nexpected: %q", s, x)
				}
			}
		})
	}
}

func newOutput() *bulog.Output {
	return &bulog.Output{
		Levels:   []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: "INFO",
		Writer:   new(bytes.Buffer),
	}
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
