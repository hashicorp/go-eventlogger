package main

import (
	"github.com/hashicorp/eventlogger/cmd/classtaglint/classtagcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(classtagcheck.Analyzer)
}
