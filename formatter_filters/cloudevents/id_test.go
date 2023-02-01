// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudevents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewId(t *testing.T) {
	t.Parallel()
	assert, require := assert.New(t), require.New(t)
	got, err := newId()
	require.NoError(err)
	assert.Len(got, 10)
}
