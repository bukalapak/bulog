package bulog

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"sync"
)

type Output struct {
	Levels     []string
	MinLevel   string
	Writer     io.Writer
	skipLevels map[string]struct{}
	pattern    *regexp.Regexp
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
		if strings.EqualFold(string(l), string(w.MinLevel)) {
			break
		}

		levels[strings.ToUpper(string(l))] = struct{}{}
	}

	w.skipLevels = levels
	w.pattern = regexp.MustCompile(`(?P<key>\w+)=(?P<value>[^\s]+)`)
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

func (w *Output) formatLine(level string, line []byte) []byte {
	return w.formatLineNone(level, line)
}

func (w *Output) formatLineNone(level string, line []byte) []byte {
	if !w.pattern.Match(line) {
		return append([]byte("["+level+"] "), line...)
	}

	var b bytes.Buffer

	b.WriteString("[" + level + "]")

	for _, submatches := range w.pattern.FindAllSubmatch(line, -1) {
		b.WriteString(" ")
		b.Write(submatches[0])
	}

	b.WriteString(" ")
	b.Write(w.extractMessage(line))
	b.WriteString("\n")

	return b.Bytes()
}

func (w *Output) extractMessage(line []byte) []byte {
	return bytes.TrimSpace(w.pattern.ReplaceAll(line, []byte("")))
}
