package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	//"cr-webserver/apiserver/util"
	"nightwatch/actions"
	"nightwatch/filters"
	"nightwatch/probes"
	"nightwatch/util/cmd"

	"github.com/golang/glog"
)

// Monitor is a unit of monitoring.
//
// It consists of a (configured) probe, zero or one filter, and one or
// more actions.  cr-monitor will invoke Prover.Probe periodically at given
// interval.
type Monitor struct {
	id       int
	name     string
	probe    probes.Prober
	filter   filters.Filter
	actors   []actions.Actor
	interval time.Duration
	timeout  time.Duration
	min      float64
	max      float64
	failedAt *time.Time

	//Status
	status string
	times  int64

	// goroutine management
	lock sync.Mutex
	env  *cmd.Environment
}

// NewMonitor creates and initializes a monitor.
//
// name can be any descriptive string for the monitor.
// p and a should not be nil.  f may be nil.
// interval is the interval between probes.
// timeout is the maximum duration for a probe to run.
// min and max defines the range for normal probe results.
func NewMonitor(
	name string,
	p probes.Prober,
	f filters.Filter,
	a []actions.Actor,
	interval, timeout time.Duration,
	min, max float64) *Monitor {
	return &Monitor{
		id:       uninitializedID,
		name:     name,
		probe:    p,
		filter:   f,
		actors:   a,
		interval: interval,
		timeout:  timeout,
		min:      min,
		max:      max,
		times:    0,
		status:   "running",
	}
}

// Start starts monitoring.
// If already started, this returns a non-nil error.
func (m *Monitor) Start() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.env != nil {
		return ErrStarted
	}

	m.env = cmd.NewEnvironment(context.Background())
	m.env.Go(m.run)

	glog.Infof("monitor started, monitor: %s", m.name)

	return nil
}

// Stop stops monitoring.
func (m *Monitor) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.env == nil {
		return
	}

	glog.Infof("monitor is stopping, monitor: %s", m.name)

	m.env.Cancel(nil)
	m.env.Wait()
	m.env = nil

	m.failedAt = nil
	m.status = "stopped"

	glog.Infof("monitor stopped, monitor: %s", m.name)
}

func (m *Monitor) die() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.env = nil
}

func callProbe(ctx context.Context, p probes.Prober, timeout time.Duration) float64 {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return p.Probe(ctx)
}

func (m *Monitor) run(ctx context.Context) error {
	if m.filter != nil {
		m.filter.Init()
	}
	for _, a := range m.actors {
		err := a.Init(m.name)
		if err != nil {
			glog.Errorf("failed to init action, monitor: %s, action: %s", m.name, a.String())
			m.die()
			return err
		}
	}

	for {
		// create a timer before starting probe.
		// This way, we can keep consistent interval between probes.
		t := time.After(m.interval)

		glog.Infof("Switch to monitor: %s", m.name)
		v := callProbe(ctx, m.probe, m.timeout)
		m.times++

		// check cancel
		select {
		case <-ctx.Done():
			return nil
		default:
			// not canceled
		}

		if m.filter != nil {
			v = m.filter.Put(v)
		}

		if (v < m.min) || (m.max < v) {
			m.status = "failed"
			if m.failedAt == nil {
				now := time.Now()
				m.failedAt = &now
				for _, a := range m.actors {
					if err := a.Fail(m.name, v); err != nil {
						glog.Errorf("failed to call Actor.Fail, monitor: %s, action: %s", m.name, a.String())
					}
				}
				glog.Warningf("monitor failure, monitor: %s, value: %s", m.name, fmt.Sprint(v))
			}
		} else {
			m.status = "running"
			if m.failedAt != nil {
				d := time.Since(*m.failedAt)
				for _, a := range m.actors {
					if err := a.Recover(m.name, d); err != nil {
						glog.Errorf("failed to call Actor.Recover, monitor: %s, action: %s", m.name, a.String())
					}
				}
				m.failedAt = nil
				glog.Warningf("monitor recovery, monitor: %s, duration: %v", m.name, int(d.Seconds()))
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t:
			// interval timer expires
		}
	}
}

// ID returns the monitor ID.
//
// ID is valid only after registration.
func (m *Monitor) ID() int {
	return m.id
}

// Name returns the name of the monitor.
func (m *Monitor) Name() string {
	return m.name
}

// String is the same as Name.
func (m *Monitor) String() string {
	return m.name
}

// Failing returns true if the monitor is detecting a failure.
func (m *Monitor) Failing() bool {
	return m.failedAt != nil
}

// Running returns true if the monitor is running.
func (m *Monitor) Running() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.env != nil
}

// Status returns the status of the monitor, current status: running, pending, failed
func (m *Monitor) Status() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.status
}

func (m *Monitor) Times() int64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.times
}

func (m *Monitor) FailedAt() string {
	m.lock.Lock()
	defer m.lock.Unlock()

	failedAt := ""
	if m.failedAt != nil {
		failedAt = m.failedAt.Format("2006-01-02 15:04:05")
	}

	return failedAt
}
