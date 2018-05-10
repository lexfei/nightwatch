// Package all import all probes to be compiled-in.
package all

import (
	// import all probes
	_ "nightwatch/probes/monitorA"
	_ "nightwatch/probes/monitorB"
)
