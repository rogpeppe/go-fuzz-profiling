package lpcodecs

import (
	"fmt"
	"time"

	protocol "github.com/influxdata/line-protocol"
)

type lineProtocolDecoder struct{}

func (lineProtocolDecoder) Decode(input *DecodeInput) (*Metric, error) {
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

func fromLineProtocolMetric(m protocol.Metric) (*Metric, error) {
	m1 := &Metric{
		Time: m.Time().UnixNano(),
		Name: []byte(m.Name()),
	}
	for _, f := range m.FieldList() {
		m1.Fields = append(m1.Fields, Field{
			Key:   Bytes(f.Key),
			Value: MustNewValue(f.Value),
		})
	}
	for _, t := range m.TagList() {
		m1.Tags = append(m1.Tags, Tag{
			Key:   Bytes(t.Key),
			Value: Bytes(t.Value),
		})
	}
	return m1, nil
}
