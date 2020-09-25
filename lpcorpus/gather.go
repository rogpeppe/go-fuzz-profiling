package lpcorpus

import (
	"sync"
	"time"
)

type DecodeInput struct {
	Text        Bytes     `yaml:"text"`
	DefaultTime int64     `yaml:"defaultTime"`
	Precision   Precision `yaml:"precision"`
}

type EncodeInput struct {
	Metric            *Metric   `yaml:"metric"`
	OmitInvalidFields bool      `yaml:"omitInvalidFields"`
	UintSupport       bool      `yaml:"uintSupport"`
	Precision         Precision `yaml:"precision"`
}

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

var (
	mu           sync.Mutex
	decodeInputs []DecodeInput
	encodeInputs []EncodeInput
)

func AddDecodeInput(t DecodeInput) {
	mu.Lock()
	defer mu.Unlock()
	decodeInputs = append(decodeInputs, t)
}

func AddEncodeInput(t EncodeInput) {
	mu.Lock()
	defer mu.Unlock()
	encodeInputs = append(encodeInputs, t)
}
