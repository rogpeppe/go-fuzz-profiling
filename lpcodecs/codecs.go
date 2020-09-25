// Package lpcodecs contains various Go implementations of line-protocol
// codecs shoehorned into a consistent interface.
package lpcodecs

import (
	"fmt"
)

type Implementation struct {
	Decoder Decoder
}

var Implementations = map[string]Implementation{
	"lineprotocol": {
		Decoder: lineProtocolDecoder{},
	},
}

type Decoder interface {
	Decode(input *DecodeInput) (*Metric, error)
}

// SkipError is returned from Encoder or Decoder when
// the encode or decode operation isn't supported
// for whatever reason.
type SkipError struct {
	Reason string
}

func (err *SkipError) Error() string {
	return fmt.Sprintf("skipped: %v", err.Reason)
}

// SkipErrorf returns a *SkipError instance with the given
// reason (formatted as for fmt.Sprintf).
func SkipErrorf(f string, a ...interface{}) error {
	return &SkipError{
		Reason: fmt.Sprintf(f, a...),
	}
}
