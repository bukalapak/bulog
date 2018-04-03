package bulog_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/bukalapak/bulog"
)

func TestWriter(t *testing.T) {
	b := new(bytes.Buffer)
	w := &bulog.Output{
		Levels:   []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: "INFO",
		Writer:   b,
	}

	l := log.New(w, "", 0)
	l.Println("info")
	l.Println("[TRACE] trace")
	l.Println("[DEBUG] debug")
	l.Println("[INFO] info")
	l.Println("[warn] warning")
	l.Println("[ERROR] error")

	s := b.String()
	x := "[INFO] info\n[INFO] info\n[WARN] warning\n[ERROR] error\n"

	if s != x {
		t.Fatalf("\nactual: %s\nexpected: %s", s, x)
	}
}
