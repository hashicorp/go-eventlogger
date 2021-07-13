package example

import (
	"time"
)

type example struct {
	ExampleX               string    `json:"name,omitempty" xml:"name"`
	ExampleY               string    `json:" xml:"name"`
	ExampleZ               string    `""`
	ExampleZz              string    ``
	PublicId               string    `class:"public"`
	SensitiveUserName      string    `class:"sensitive"`
	SensitiveUserNameOops1 string    `class:"sensitve"`                    // want "invalid data classification: \"sensitve\""
	SensitiveUserNameOops2 string    `class:"sensitive,secrt"`             // want "invalid filter operation: \"secrt\""
	SensitiveBad           string    `class:"senitive"`                    // want "invalid data classification: \"senitive\""
	SensitiveDouble        string    `class:"sensitive" class:"sensitive"` // want "found 2 data classifications for single field"
	SensitiveNonFilterable time.Time `class:"sensitive"`                   // want "invalid data classification for non-filterable type"
	NotFiltered            string    `class:"public,redact"`               // want "filter operations invalid on public data classifications"
	NotFilteredEither      string    `class:"public,redct"`                // want "filter operations invalid on public data classifications" "invalid filter operation: \"redct\""
	More                   string    `class:"sensitive,redact,more"`       // want "too many classification options given: 3"
	NonFilterable          time.Time `class:"public"`
	LoginTimestamp         time.Time
}
