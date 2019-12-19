package imap

import (
	"strings"
)

// Path is a path in an IMap.
type Path struct {
	Keys []string
}

func NewPath(keys ...string) *Path {
	return &Path{
		Keys: keys,
	}
}

func (p *Path) String() string {
	var str strings.Builder
	for i, key := range p.Keys {
		if i > 0 {
			str.WriteString("/")
		}
		str.WriteString(key)
	}
	return str.String()
}
