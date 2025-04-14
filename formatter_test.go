// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"context"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"
)

func TestJSONFormatter(t *testing.T) {
	w := &JSONFormatter{}
	e := &Event{
		Formatted: make(map[string][]byte),
	}
	_, err := w.Process(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProtoFormatter(t *testing.T) {
	w := &ProtoFormatter{}
	e := &Event{
		Formatted: make(map[string][]byte),
		Payload: &anypb.Any{
			TypeUrl: "example.com/test.ExampleEvent",
			Value:   []byte("test"),
		},
	}
	_, err := w.Process(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}
}
