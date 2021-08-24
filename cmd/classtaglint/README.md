# `classtaglint` 

The `classtaglint` command implements a [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis)
based linter to analyze usage of the [`github.com/hashicorp/eventlogger/filters/encrypt`](https://pkg.go.dev/github.com/hashicorp/eventlogger/filters/encrypt)'s
`class` struct tags.

## Usage

```console
$ classtaglint
classtaglint: analyze usage of class struct tags

Usage: classtaglint [-flag] [package]


Flags:
  -V    print version and exit
  -all
        no effect (deprecated)
  -c int
        display offending line with this many lines of context (default -1)
  -cpuprofile string
        write CPU profile to this file
  -debug string
        debug flags, any subset of "fpstv"
  -fix
        apply all suggested fixes
  -flags
        print analyzer flags in JSON
  -json
        emit JSON output
  -memprofile string
        write memory profile to this file
  -source
        no effect (deprecated)
  -tags string
        no effect (deprecated)
  -trace string
        write trace log to this file
  -v    no effect (deprecated)
```

```
$ cd classtagcheck/testdata/src/a/
classtaglint .
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:14:35: invalid data classification: "sensitve"
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:15:35: invalid filter operation: "secrt"
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:16:35: invalid data classification: "senitive"
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:17:35: found 2 data classifications for single field
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:18:35: invalid data classification for non-filterable type
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:19:35: filter operations invalid on public data classifications
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:20:35: filter operations invalid on public data classifications
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:20:35: invalid filter operation: "redct"
/full/path/go-eventlogger/cmd/classtaglint/classtagcheck/testdata/src/a/main.go:21:35: too many classification options given: 3
$ classtaglint -json .
...
$ go vet -vettool=$(which classtaglint)
...
$ go vet -vettool=$(which classtaglint) -json 
```

