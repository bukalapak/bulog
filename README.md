# bulog

[![Build Status](https://travis-ci.org/bukalapak/bulog.svg?branch=master)](https://travis-ci.org/bukalapak/bulog)
[![codecov](https://codecov.io/gh/bukalapak/bulog/branch/master/graph/badge.svg)](https://codecov.io/gh/bukalapak/bulog)
[![GoDoc](https://godoc.org/github.com/bukalapak/bulog?status.svg)](https://godoc.org/github.com/bukalapak/bulog)
[![Go Report Card](https://goreportcard.com/badge/github.com/bukalapak/bulog)](https://goreportcard.com/report/github.com/bukalapak/bulog)

An alternative output destination for the standard library log package

## Usage

When your application use default `log` package.

```go
log.Println("hello world!")

// 2018/04/04 11:07:36 hello world!
```

You can enhance it a bit using `bulog.Standard`:

```go
l := bulog.Standard(os.Stderr)
l.Println("hello world!")

// {"@timestamp":"2018-10-22T21:02:43+07:00","message":"hello world!"}
```

You can embed log level within the message:

```go
l := bulog.Standard(os.Stderr)
l.Println("[INFO] hello world!")

// {"level":"info","@timestamp":"2018-10-22T21:04:27+07:00","message":"hello world!"}
```

If you prefer writing log message using [logfmt](https://brandur.org/logfmt) format, bulog provides the support via `bulog.Logfmt`:

```go
l := bulog.Logfmt(os.Stdout)
l.Println(`level="info" code=200 msg="hello world!"`)

// {"level":"info","code":200,"@timestamp":"2018-10-22T21:07:45+07:00","message":"hello world!"}
```

That's it.
