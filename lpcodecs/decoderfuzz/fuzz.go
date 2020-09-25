package decoderfuzz

import (
	"fmt"
	"time"

	"github.com/rogpeppe/line-protocol-corpus/lpcodecs"
	"github.com/rogpeppe/line-protocol-corpus/lpcorpus"
)

var defaultTime = time.Date(2000, 1, 2, 12, 13, 14, 0, time.UTC).UnixNano()

const (
	normalExit       = 0
	highPriorityExit = 1
	noAddCorpusExit  = -1
)

func Fuzz(data []byte) int {
	input := &lpcorpus.DecodeInput{
		Text:        data,
		DefaultTime: defaultTime,
		Precision: lpcorpus.Precision{
			Duration: time.Nanosecond,
		},
	}
	exitCode := normalExit
	outputs := make(map[string]*lpcorpus.DecodeOutput)
	for name, impl := range lpcodecs.Implementations {
		if name != "lineprotocol" {
			continue
		}
		m, err := impl.Decoder.Decode(input)
		if err != nil {
			if _, ok := err.(*lpcodecs.SkipError); ok {
				continue
			}
			outputs[name] = &lpcorpus.DecodeOutput{
				Error: err.Error(),
			}
		} else {
			exitCode = highPriorityExit
			outputs[name] = &lpcorpus.DecodeOutput{
				Result: m,
			}
		}
	}
	if len(outputs) < 2 {
		return exitCode
	}
	var o *lpcorpus.DecodeOutput
	var firstName string
	for name, d := range outputs {
		if o == nil {
			firstName = name
			o = d
			continue
		}
		if !d.Equal(o) {
			panic(fmt.Sprintf("inconsistent result between %v and %v", firstName, name))
		}
	}
	return exitCode
}
