// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import "errors"

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrNodeNotFound     = errors.New("node not found")
)
