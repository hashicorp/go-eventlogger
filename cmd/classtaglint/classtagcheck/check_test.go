package classtagcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata string

func init() {
	testdata = analysistest.TestData()
}

func TestExample(t *testing.T) {
	analysistest.Run(t, testdata, Analyzer, "example")
}
