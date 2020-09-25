package lpcodecs

import (
	"bytes"
	"fmt"
	"time"

	protocol "github.com/influxdata/line-protocol"
	"github.com/rogpeppe/line-protocol-corpus/lpcorpus"
)

type lineProtocolDecoder struct{}

func (lineProtocolDecoder) Decode(input *lpcorpus.DecodeInput) (*lpcorpus.Metric, error) {
	h := protocol.NewMetricHandler()
	h.SetTimePrecision(input.Precision.Duration)
	h.SetTimeFunc(func() time.Time {
		return time.Unix(0, input.DefaultTime)
	})
	parser := protocol.NewParser(h)
	ms, err := parser.Parse(input.Text)
	if err != nil {
		return nil, err
	}
	if len(ms) != 1 {
		return nil, fmt.Errorf("unexpected number of points (got %d want 1)", len(ms))
	}
	m, err := fromLineProtocolMetric(ms[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert metric: %v", err)
	}
	return m, nil
}

type lineProtocolEncoder struct{}

func (lineProtocolEncoder) Encode(input *lpcorpus.EncodeInput) ([]byte, error) {
	var buf bytes.Buffer
	enc := protocol.NewEncoder(&buf)
	var typeSupport protocol.FieldTypeSupport
	if input.UintSupport {
		typeSupport |= protocol.UintSupport
	}
	enc.SetFieldTypeSupport(typeSupport)
	if !input.OmitInvalidFields {
		enc.FailOnFieldErr(true)
	}
	enc.SetPrecision(input.Precision.Duration)
	if _, err := enc.Encode(protocolMetric{input.Metric}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fromLineProtocolMetric(m protocol.Metric) (*lpcorpus.Metric, error) {
	m1 := &lpcorpus.Metric{
		Time: m.Time().UnixNano(),
		Name: []byte(m.Name()),
	}
	for _, f := range m.FieldList() {
		m1.Fields = append(m1.Fields, lpcorpus.Field{
			Key:   lpcorpus.Bytes(f.Key),
			Value: lpcorpus.MustNewValue(f.Value),
		})
	}
	for _, t := range m.TagList() {
		m1.Tags = append(m1.Tags, lpcorpus.Tag{
			Key:   lpcorpus.Bytes(t.Key),
			Value: lpcorpus.Bytes(t.Value),
		})
	}
	return m1, nil
}

// protocolMetric implements protocol.Metric on a lpcorpus.Metric.
type protocolMetric struct {
	*lpcorpus.Metric
}

func (m protocolMetric) Name() string {
	return string(m.Metric.Name)
}

func (m protocolMetric) Time() time.Time {
	return time.Unix(0, m.Metric.Time)
}

func (m protocolMetric) TagList() []*protocol.Tag {
	tags := make([]*protocol.Tag, len(m.Tags))
	for i, tag := range m.Tags {
		tags[i] = &protocol.Tag{
			Key:   string(tag.Key),
			Value: string(tag.Value),
		}
	}
	return tags
}

func (m protocolMetric) FieldList() []*protocol.Field {
	fields := make([]*protocol.Field, len(m.Fields))
	for i, field := range m.Fields {
		fields[i] = &protocol.Field{
			Key:   string(field.Key),
			Value: field.Value.Interface(),
		}
	}
	return fields
}
