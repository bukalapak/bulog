package bulog_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/bukalapak/bulog"
)

func TestOutput(t *testing.T) {
	m := map[string][][]string{
		"AutoLevel": [][]string{
			[]string{"info", "[INFO] info"},
			[]string{"[INFO] info", "[INFO] info"},
		},
		"NormalizeLevel": [][]string{
			[]string{"[warn] warning", "[WARN] warning"},
			[]string{"[WARN] warning", "[WARN] warning"},
		},
		"SkipLevel": [][]string{
			[]string{"[INFO] info", "[DEBUG] debug"},
			[]string{"[INFO] info"},
		},
		"WithMetadata": [][]string{
			[]string{`[INFO] info foo="bar" num=8`},
			[]string{`[INFO] info foo=bar num=8`},
		},
	}

	for k, v := range m {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = bulog.None
			l := log.New(w, "", 0)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[1], "\n") + "\n"

			if s != x {
				t.Fatalf("\nactual: %q\nexpected: %q", s, x)
			}
		})
	}
}

func TestOutput_Fmt(t *testing.T) {
	m := map[string][][]string{
		"SkipLevel": [][]string{
			[]string{"[INFO] info", "[DEBUG] debug"},
			[]string{`level=INFO msg=info`},
		},
		"Output": [][]string{
			[]string{`[INFO] hello info num=8 foo="bar baz"`},
			[]string{`level=INFO num=8 foo="bar baz" msg="hello info"`},
		},
		"WithMetadata": [][]string{
			[]string{`[INFO] info foo="bar" num=8`},
			[]string{`level=INFO foo=bar num=8 msg=info`},
		},
		"MsgMetadata": [][]string{
			[]string{`[INFO] foo="bar" num=8 msg="info"`},
			[]string{`level=INFO foo=bar num=8 msg=info`},
		},
		"SpaceMetadata": [][]string{
			[]string{`[INFO] foo="bar baz" info`},
			[]string{`level=INFO foo="bar baz" msg=info`},
		},
	}

	for k, v := range m {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = bulog.Logfmt
			l := log.New(w, "", 0)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[1], "\n") + "\n"

			if s != x {
				t.Fatalf("\nactual: %q\nexpected: %q", s, x)
			}
		})
	}
}

func TestOutput_JSON(t *testing.T) {
	m := map[string][][]string{
		"SkipLevel": [][]string{
			[]string{"[INFO] info", "[DEBUG] debug"},
			[]string{`{"level":"INFO","msg":"info"}`},
		},
		"WithMetadata": [][]string{
			[]string{`[INFO] info foo="bar" num=8 float=9.99 bool=true`},
			[]string{`{"level":"INFO","foo":"bar","num":8,"float":9.99,"bool":true,"msg":"info"}`},
		},
		"MsgMetadata": [][]string{
			[]string{`[INFO] foo="bar" num=8 msg="info"`},
			[]string{`{"level":"INFO","foo":"bar","num":8,"msg":"info"}`},
		},
		"SpaceMetadata": [][]string{
			[]string{`[INFO] foo="bar baz" info`},
			[]string{`{"level":"INFO","foo":"bar baz","msg":"info"}`},
		},
	}

	for k, v := range m {
		t.Run(k, func(t *testing.T) {
			w := newOutput()
			w.Format = bulog.JSON
			l := log.New(w, "", 0)

			for i := range v[0] {
				l.Println(v[0][i])
			}

			s := w.Writer.(*bytes.Buffer).String()
			x := strings.Join(v[1], "\n") + "\n"

			if s != x {
				t.Fatalf("\nactual: %q\nexpected: %q", s, x)
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
