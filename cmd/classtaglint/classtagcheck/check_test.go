package classtagcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata string

func init() {
	// NOTE: a part of of github.com/hashicorp/eventlogger/filters/encrypt and
	//       google.golang.org/protobuf/types/known needed to be copied or mocked
	//       inside of the testdata/src directory so that it could be loaded by
	//       analysistest. The analysistest package is not module-aware.
	testdata = analysistest.TestData()
}

func TestExample(t *testing.T) {
	analysistest.Run(t, testdata, Analyzer, "example")
}

func TestExampleMapTags(t *testing.T) {
	analysistest.Run(t, testdata, Analyzer, "example_map_tags")
}

func TestExampleMapTagsComplex(t *testing.T) {
	analysistest.Run(t, testdata, Analyzer, "example_map_tags_complex")
}

func TestExampleStructPBTags(t *testing.T) {
	analysistest.Run(t, testdata, Analyzer, "example_structpb_tags")
}
