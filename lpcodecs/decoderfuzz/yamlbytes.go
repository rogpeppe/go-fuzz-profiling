package decoderfuzz

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Bytes implements a byte slice that knows how to marshal itself into
// and out of YAML while preserving raw-byte integrity by using
// YAML tags. If the byte slice is valid UTF-8, it is marshaled as a string.
type Bytes []byte

func (b Bytes) MarshalYAML() (interface{}, error) {
	// Note: work around cue issue with importing zero bytes too.
	if utf8.Valid(b) && bytes.IndexByte(b, 0) == -1 {
		if len(b) > 0 && b[0] == '\t' {
			// Work around https://github.com/go-yaml/yaml/issues/660
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Style: yaml.SingleQuotedStyle,
				Value: string(b),
			}, nil
		}
		return string(b), nil
	}
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!binary",
		Style: yaml.SingleQuotedStyle,
		Value: base64.StdEncoding.EncodeToString(b),
	}, nil
}

func (b *Bytes) UnmarshalYAML(n *yaml.Node) error {
	switch n.Tag {
	case "!!str":
		*b = []byte(n.Value)
	case "!!binary":
		data, err := base64.StdEncoding.DecodeString(n.Value)
		if err != nil {
			return fmt.Errorf("invalid base64 value in binary-tagged value: %v", err)
		}
		*b = data
	default:
		return fmt.Errorf("cannot unmarshal %s `%s` into Bytes", n.Tag, n.Value)
	}
	return nil
}
