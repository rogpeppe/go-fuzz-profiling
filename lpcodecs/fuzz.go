package lpcodecs

import (
	"fmt"
	"time"

	protocol "github.com/influxdata/line-protocol"
)

var defaultTime = time.Date(2000, 1, 2, 12, 13, 14, 0, time.UTC).UnixNano()

func Fuzz(data []byte) int {
	exitCode := 0
	_, err := lineProtocolDecode(data)
	if err == nil {
		exitCode = 1
	}
	return exitCode
}

func lineProtocolDecode(input []byte) (protocol.Metric, error) {
	h := protocol.NewMetricHandler()
	h.SetTimePrecision(time.Nanosecond)
	h.SetTimeFunc(func() time.Time {
		return time.Unix(0, defaultTime)
	})
	parser := protocol.NewParser(h)
	ms, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}
	if len(ms) != 1 {
		return nil, fmt.Errorf("unexpected number of points (got %d want 1)", len(ms))
	}
	return ms[0], nil
}
