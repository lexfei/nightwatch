package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"nightwatch"
	_ "nightwatch/actions/all"
	_ "nightwatch/filters/all"
	"nightwatch/monitor"
	_ "nightwatch/probes/all"
	"nightwatch/util/cmd"
	"nightwatch/util/version"

	"github.com/golang/glog"
)

const (
	defaultConfDir    = "/usr/local/etc/nightwatch"
	defaultListenAddr = "localhost:3838"
)

var (
	confDir    = flag.String("c", defaultConfDir, "directory for monitor configs")
	listenAddr = flag.String("s", defaultListenAddr, "HTTP server address")
	vinfo      = flag.Bool("version", false, "show version info.")
)

func usage() {
	fmt.Fprint(os.Stderr, `Usage: nightwatch [options] COMMAND [arg...]

If COMMAND is "server", nightwatch runs in server mode.
For other commands, nightwatch works as a client for nightwatch server.

Options:
`)
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, `
Commands:
    server              Start agent server.
    list               List registered monitors.
    register FILE      Register monitors defined in FILE.
                       If FILE is "-", nightwatch reads from stdin.
    show ID            Show the status of a monitor for ID.
    start ID           Start a monitor.
    stop ID            Stop a monitor.
    unregister ID      Stop and unregister a monitor.
    verbosity [LEVEL]  Query or change logging threshold.
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if *vinfo {
		v := version.Get()
		marshalled, err := json.MarshalIndent(&v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(marshalled))
		return
	}

	//cmd.LogConfig{}.Apply()

	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	command := args[0]

	if command != "server" {
		err := runCommand(command, args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, strings.TrimSpace(err.Error()))
			os.Exit(1)
		}
		return
	}

	if err := loadConfigs(*confDir); err != nil {
		glog.Errorf("loadConfigs failed!error: %v", err)
		os.Exit(1)
	}

	nightwatch.Server(*listenAddr)
	glog.Infof("nightwatch listen on: %s", *listenAddr)
	err := cmd.Wait()
	if err != nil && !cmd.IsSignaled(err) {
		glog.Errorf("nightwatch encounter unknow abnormal!error: %v", err)
		os.Exit(1)
	}

	// stop all monitors gracefully.
	for _, m := range monitor.ListMonitors() {
		glog.Infof("stop monitor: %s", m.String())
		m.Stop()
	}
}
