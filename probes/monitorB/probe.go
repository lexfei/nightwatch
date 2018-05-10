package monitorB

import (
	"context"
	"fmt"
	"time"

	"nightwatch/probes"
)

const (
	DELETING_TIMEOUT_DURATION = 1 * time.Hour
)

type probe struct{}

func init() {
	probes.Register("monitorB", construct)
}

func (p *probe) String() string {
	return "probe:monitorB:monitorB"
}

func construct(params map[string]interface{}) (probes.Prober, error) {
	return &probe{}, nil
}

func (p *probe) Probe(ctx context.Context) float64 {
	fmt.Printf("do some deal with monitor: %s\n", p.String())
	return 0
}
