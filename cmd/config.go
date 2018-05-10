package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"nightwatch"
	"nightwatch/monitor"

	"github.com/ghodss/yaml"
)

func loadYAML(f string) ([]*nightwatch.MonitorDefinition, error) {
	s := []*struct {
		Monitor *nightwatch.MonitorDefinition `yaml:"monitor"`
	}{nil}

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	fn := func() error {
		return yaml.Unmarshal(content, &s)
	}

	if f == "-" {
		fn = func() error {
			content, _ := ioutil.ReadAll(os.Stdin)
			return yaml.Unmarshal(content, &s)
		}
	}
	err = fn()
	if err != nil {
		return nil, err
	}
	monitors := make([]*nightwatch.MonitorDefinition, 0)
	for _, monitor := range s {
		monitors = append(monitors, monitor.Monitor)
	}

	return monitors, nil
}

func loadFile(f string) error {
	defs, err := loadYAML(f)
	if err != nil {
		return err
	}

	monitors := make([]*monitor.Monitor, 0, len(defs))
	for _, md := range defs {
		m, err := nightwatch.CreateMonitor(md)
		if err != nil {
			return err
		}
		monitors = append(monitors, m)
	}

	for _, m := range monitors {
		// ignoring errors is safe at this point.
		monitor.Register(m)
		m.Start()
	}
	return nil
}

func loadConfigs(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := loadFile(f); err != nil {
			return err
		}
	}
	return nil
}
