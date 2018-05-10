package nightwatch

import (
	"net/http"
	"time"

	"nightwatch/util/cmd"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second

	// Version may be used for REST API version checks in future.
	Version = "1.0"

	// VersionHeader is the HTTP request header for Version.
	VersionHeader = "X-Watcher-Version"
)

// Serve runs REST API server until the global environment is canceled.
func Server(addr string) error {
	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:         addr,
			Handler:      NewRouter(),
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
		},
	}
	return s.ListenAndServe()
}
