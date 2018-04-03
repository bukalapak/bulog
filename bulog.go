package bulog

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

type Output struct {
	Levels     []string
	MinLevel   string
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

	return w.Writer.Write(normalizeLine(l, z))
}

func (w *Output) init() {
	levels := make(map[string]struct{})
	for _, l := range w.Levels {
		if strings.EqualFold(string(l), string(w.MinLevel)) {
			break
		}

		levels[strings.ToUpper(string(l))] = struct{}{}
	}

	w.skipLevels = levels
}

func (w *Output) extractLine(line []byte) (string, []byte) {
	level := extractLevel(line)

	if level != "" {
		return level, line[(len(level) + 3):]
	}

	return w.MinLevel, line
}

func extractLevel(line []byte) (level string) {
	x := bytes.IndexByte(line, '[')
	if x >= 0 {
		y := bytes.IndexByte(line[x:], ']')

		if y >= 0 {
			level = string(bytes.ToUpper(line[x+1 : x+y]))
		}
	}

	return
}

func normalizeLine(level string, line []byte) []byte {
	return append([]byte("["+level+"] "), line...)
}
