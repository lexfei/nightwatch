package monitorA

import (
	"context"
	"fmt"
	"time"

	"nightwatch"
	"nightwatch/probes"
)

const (
	DELETING_TIMEOUT_DURATION = 1 * time.Hour
)

type probe struct {
	duration int // hour, 检查几小时之内的
}

func init() {
	probes.Register("monitorA", construct)
}

func (p *probe) String() string {
	return "probe:monitorA:monitorA"
}

func construct(params map[string]interface{}) (probes.Prober, error) {
	duration, err := nightwatch.GetInt("duration", params)
	if err != nil {
		return nil, err
	}
	return &probe{
		duration: duration,
	}, nil
}

func (p *probe) Probe(ctx context.Context) float64 {
	fmt.Printf("do some deal with monitor: %s\n", p.String())
	return 0
}
