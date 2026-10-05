// Copyright IBM Corp. 2019, 2026
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
