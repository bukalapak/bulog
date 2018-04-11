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
w.LogFlags = log.Flags()
w.LogPrefix = log.Prefix()

log.SetOutput(w)
log.Println("hello world!")

// level=INFO msg="hello world!" caller=/home/subosito/src/github.com/subosito/playground/main.go:13 stacktrace="goroutine 1 [running]:\nruntime/debug.Stack(0x0, 0xc42001c100, 0xc420094000)\n\t/usr/lib/go/src/runtime/debug/stack.go:24 +0xa7\ngithub.com/bukalapak/bulog.stacktrace(0x4bcf60, 0xc4200741e0, 0x4d857a)\n\t/home/subosito/src/github.com/bukalapak/bulog/bulog.go:222 +0x22\ngithub.com/bukalapak/bulog.(*Output).parseLine(0xc420084000, 0xc420014170, 0xd, 0x10, 0x4acf479bbbef14b0)\n\t/home/subosito/src/github.com/bukalapak/bulog/bulog.go:144 +0x5a0\ngithub.com/bukalapak/bulog.(*Output).formatLineLogfmt(0xc420084000, 0x4d8279, 0x4, 0xc420014170, 0xd, 0x10, 0xc420088000, 0x0, 0x0)\n\t/home/subosito/src/github.com/bukalapak/bulog/bulog.go:103 +0x81\ngithub.com/bukalapak/bulog.(*Output).formatLine(0xc420084000, 0x4d8279, 0x4, 0xc420014170, 0xd, 0x10, 0xc420014170, 0xd, 0x10)\n\t/home/subosito/src/github.com/bukalapak/bulog/bulog.go:97 +0x6c\ngithub.com/bukalapak/bulog.(*Output).Write(0xc420084000, 0xc420014170, 0xd, 0x10, 0xd, 0xc420014170, 0x0)\n\t/home/subosito/src/github.com/bukalapak/bulog/bulog.go:64 +0x16d\nlog.(*Logger).Output(0xc420078140, 0x2, 0xc420014160, 0xd, 0x0, 0x0)\n\t/usr/lib/go/src/log/log.go:172 +0x1fd\nlog.(*Logger).Println(0xc420078140, 0xc420059f68, 0x1, 0x1)\n\t/usr/lib/go/src/log/log.go:188 +0x6a\nmain.main()\n\t/home/subosito/src/github.com/subosito/playground/main.go:13 +0x1c5\n" timestamp=2018-04-04T11:08:51+07:00
```

Instead of using standard logger, you can also custom one, in this case, you can use `Attach` method:

```go
l := log.New(os.Stderr, "", log.LstdFlags)

w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.Attach(l)

l.Println("hello world!")
```

You can remove `stacktrace` and `caller`, also change timestamp format:

```go
l := log.New(os.Stderr, "", log.LstdFlags)

w:= bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.ShowCaller = false
w.Stacktrace = false
w.TimeFormat = time.RFC822
w.Attach(l)

l.Println("hello world!")

// level=INFO msg="hello world!" timestamp="04 Apr 18 11:11 WIB"
```

Since we already defined log levels, let's use it:

```go
l := log.New(os.Stderr, "", log.LstdFlags)

w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.ShowCaller = false
w.Stacktrace = false
w.TimeFormat = time.RFC822
w.Attach(l)

l.Println("hello world!")
l.Println("[WARN] warning message")
l.Println("[DEBUG] debug message")

// level=INFO msg="hello world!" timestamp="04 Apr 18 11:12 WIB"
// level=WARN msg="warning message" timestamp="04 Apr 18 11:12 WIB"
```

As you can see, `DEBUG` message is ignored.

Okay, let's produce JSON format so it can be consume by third-party tools:

```go
l := log.New(os.Stderr, "", log.LstdFlags)

w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.ShowCaller = false
w.Stacktrace = false
w.TimeFormat = time.RFC822
w.Format = bulog.JSON
w.Attach(l)

l.Println("hello world!")
l.Println("[WARN] warning message")
l.Println("[DEBUG] debug message")

// {"level":"INFO","timestamp":"04 Apr 18 11:14 WIB","msg":"hello world!"}
// {"level":"WARN","timestamp":"04 Apr 18 11:14 WIB","msg":"warning message"}
```

Put `caller` and default time format back, also change key names for level and timestamp:

```go
l := log.New(os.Stderr, "", log.LstdFlags)

w := bulog.New("INFO", []string{"DEBUG", "INFO", "WARN"})
w.Stacktrace = false
w.Format = bulog.JSON
w.KeyNames = map[string]string{
	"level":     "severity",
	"timestamp": "@timestamp",
}
w.Attach(l)

l.Println("hello world!")
l.Println("[WARN] warning message")
l.Println("[DEBUG] debug message")

// {"severity":"INFO","@timestamp":"2018-04-04T11:14:51+07:00","caller":"/home/subosito/src/github.com/subosito/playground/main.go:14","msg":"hello world!"}
// {"severity":"WARN","msg":"warning message","@timestamp":"2018-04-04T11:14:51+07:00","caller":"/home/subosito/src/github.com/subosito/playground/main.go:15"}
```

That's it.
