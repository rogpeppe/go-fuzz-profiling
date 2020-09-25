package lpcodecs

import (
	"reflect"
	"time"
)

type Metric struct {
	Time   int64   `yaml:"time"`
	Name   Bytes   `yaml:"name"`
	Tags   []Tag   `yaml:"tags,omitempty"`
	Fields []Field `yaml:"fields,omitempty"`
}

type Tag struct {
	Key   Bytes `yaml:"key"`
	Value Bytes `yaml:"value"`
}

type Field struct {
	Key   Bytes `yaml:"key"`
	Value Value `yaml:"value"`
}

type DecodeInput struct {
	Text        Bytes     `yaml:"text"`
	DefaultTime int64     `yaml:"defaultTime"`
	Precision   Precision `yaml:"precision"`
}

type DecodeOutput struct {
	Result *Metric `yaml:"result,omitempty"`
	Error  string  `yaml:"error,omitempty"`
}

// Two decode outputs compare equal if they both succeed
// with the same decoded metric or they both failed.
func (o1 *DecodeOutput) Equal(o2 *DecodeOutput) bool {
	if o1.Result != nil && o2.Result != nil {
		return reflect.DeepEqual(o1.Result, o2.Result)
	}
	return o1.Error != "" && o2.Error != ""
}

type Precision struct {
	Duration time.Duration
}

func (p Precision) MarshalText() ([]byte, error) {
	return []byte(p.Duration.String()), nil
}

func (p *Precision) UnmarshalText(data []byte) error {
	d, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	p.Duration = d
	return nil
}
