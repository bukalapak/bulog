# bulog

[![Build Status](https://travis-ci.org/bukalapak/bulog.svg?branch=master)](https://travis-ci.org/bukalapak/bulog)
[![Coverage Status](https://img.shields.io/codecov/c/github/bukalapak/bulog.svg)](https://codecov.io/gh/bukalapak/bulog)
[![GoDoc](https://godoc.org/github.com/bukalapak/bulog?status.svg)](https://godoc.org/github.com/bukalapak/bulog)

An alternative output destination for the standard library log package

## Usage

When your application use default `log` package.

```go
log.Println("hello world!")

// 2018/04/04 11:07:36 hello world!
```

You can enhance it a bit using `bulog` (with `logfmt` format):

```go
w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})

l := log.New(w, "", 0)
l.Println("hello world!")

// level=INFO msg="hello world!" stacktrace=/home/subosito/src/github.com/subosito/playground/main.go:13 timestamp=2018-04-04T11:08:51+07:00
```

You can remove `stacktrace` and change timestamp format:

```go
w:= bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.Stacktrace = false
w.TimeFormat = time.RFC822

l := log.New(w, "", 0)
l.Println("hello world!")

// level=INFO msg="hello world!" timestamp="04 Apr 18 11:11 WIB"
```

Since we already defined log levels, let's use it:

```go
w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.Stacktrace = false
w.TimeFormat = time.RFC822

l := log.New(w, "", 0)
l.Println("hello world!")
l.Println("[WARN] warning message")
l.Println("[DEBUG] debug message")

// level=INFO msg="hello world!" timestamp="04 Apr 18 11:12 WIB"
// level=WARN msg="warning message" timestamp="04 Apr 18 11:12 WIB"
```

As you can see, `DEBUG` message is ignored.

Okay, let's produce JSON format so it can be consume by third-party tools:

```go
w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.Stacktrace = false
w.TimeFormat = time.RFC822
w.Format = bulog.JSON

l := log.New(w, "", 0)
l.Println("hello world!")
l.Println("[WARN] warning message")
l.Println("[DEBUG] debug message")

// {"level":"INFO","timestamp":"04 Apr 18 11:14 WIB","msg":"hello world!"}
// {"level":"WARN","timestamp":"04 Apr 18 11:14 WIB","msg":"warning message"}
```

Put stacktrace and default time format back, also use log helper:

```go
w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.Format = bulog.JSON

l := bulog.NewLog(w)
l.Println("hello world!")
l.Println("[WARN] warning message")
l.Println("[DEBUG] debug message")

// {"level":"INFO","timestamp":"2018-04-04T11:14:51+07:00","stacktrace":"/home/subosito/src/github.com/subosito/playground/main.go:14","msg":"hello world!"}
// {"level":"WARN","msg":"warning message","timestamp":"2018-04-04T11:14:51+07:00","stacktrace":"/home/subosito/src/github.com/subosito/playground/main.go:15"}
```

That's it.
