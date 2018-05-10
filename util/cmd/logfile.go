// +build !windows

package cmd

import (
	"io"
	"syscall"

	"nightwatch/util/log"
)

func openLogFile(filename string) (io.Writer, error) {
	return log.NewFileReopener(filename, syscall.SIGUSR1)
}
