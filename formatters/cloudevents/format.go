package cloudevents

import (
	"fmt"

	"github.com/hashicorp/eventlogger"
)

const (
	DataContentTypeCloudEvents = "application/cloudevents"
	DataContentTypeText        = "text/plain"
)

type Format string

var (
	FormatJSON        Format = "cloudevents-json"
	FormatText        Format = "cloudevents-text"
	FormatUnspecified Format = ""
)

func (f Format) validate() error {
	const op = "cloudevents.(Format).validate"
	switch f {
	case FormatJSON, FormatText, FormatUnspecified:
		return nil
	default:
		return fmt.Errorf("%s: '%s' is not a valid format: %w", op, f, eventlogger.ErrInvalidParameter)
	}
}

func (f Format) convertToDataContentType() string {
	switch f {
	case FormatJSON, FormatUnspecified:
		return DataContentTypeCloudEvents
	default:
		return DataContentTypeText
	}
}
