package bulog_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/bukalapak/bulog"
	"github.com/stretchr/testify/assert"
)

func TestStandard(t *testing.T) {
	data := map[string][]string{
		"Hello world!":        []string{"", "Hello world!"},
		"[INFO] Hello world!": []string{"info", "Hello world!"},
		"[WARN] Hello world!": []string{"warn", "Hello world!"},
	}

	for k, v := range data {
		out := new(bytes.Buffer)

		l := bulog.Standard(out)
		l.Println(k)

		z := struct {
			Level     string    `json:"level"`
			Timestamp time.Time `json:"@timestamp"`
			Message   string    `json:"message"`
		}{}

		dec := json.NewDecoder(out)
		err := dec.Decode(&z)
		assert.Nil(t, err)

		assert.Equal(t, v[0], z.Level)
		assert.Equal(t, v[1], z.Message)
		assert.False(t, z.Timestamp.IsZero())
	}
}

func TestLogfmt(t *testing.T) {
	data := map[string][]string{
		`msg="Hello world!"`: []string{"", "Hello world!"},
		`level=info int=200 text="OK" bool=true double=4.0 msg="Hello world!"`: []string{"info", "Hello world!"},
	}

	for k, v := range data {
		out := new(bytes.Buffer)

		l := bulog.Logfmt(out)
		l.Println(k)

		z := struct {
			Level     string    `json:"level"`
			Timestamp time.Time `json:"@timestamp"`
			Message   string    `json:"message"`
			Int       int       `json:"int"`
			Text      string    `json:"text"`
			Bool      bool      `json:"bool"`
			Double    float64   `json:"double"`
		}{}

		dec := json.NewDecoder(out)
		err := dec.Decode(&z)
		assert.Nil(t, err)

		assert.Equal(t, v[0], z.Level)
		assert.Equal(t, v[1], z.Message)
		assert.False(t, z.Timestamp.IsZero())

		if z.Level == "info" {
			assert.Equal(t, 200, z.Int)
			assert.Equal(t, 4.0, z.Double)
			assert.Equal(t, true, z.Bool)
			assert.Equal(t, "OK", z.Text)
		}
	}
}

func BenchmarkLogStd(b *testing.B) {
	l := log.New(ioutil.Discard, "INFO: ", log.LstdFlags)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		l.Println("information")
	}
}

func BenchmarkStandard(b *testing.B) {
	out := new(bytes.Buffer)
	l := bulog.Standard(out)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		l.Println("[INFO] information")
	}
}

func BenchmarkLogfmt(b *testing.B) {
	out := new(bytes.Buffer)
	l := bulog.Logfmt(out)

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		l.Println(`level="info" msg="information"`)
	}
}
