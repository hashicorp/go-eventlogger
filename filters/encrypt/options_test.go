package encrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_getOpts provides unit tests for getOpts and all the options
func Test_getOpts(t *testing.T) {
	t.Parallel()
	t.Run("withFilterOperations", func(t *testing.T) {
		assert := assert.New(t)
		filters := map[DataClassification]FilterOperation{
			UnknownClassification: RedactOperation,
			PublicClassification:  NoOperation,
			SecretClassification:  EncryptOperation,
		}
		opts := getOpts(withFilterOperations(filters))
		testOpts := getDefaultOptions()
		testOpts.withFilterOperations = filters
		assert.Equal(opts, testOpts)
	})
}
