// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
