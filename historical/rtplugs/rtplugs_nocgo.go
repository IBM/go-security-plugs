//go:build !cgo
// +build !cgo

package rtplugs

import (
	"os"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

// load() will attempt to dynamically load plugs
//
// env RTPLUGS_SO defines a comma seperated list of .so plug files
// relative/full path may be used
func load() {
	soPlugsStr := os.Getenv("RTPLUGS_SO")
	if soPlugsStr == "" {
		return
	}

	pi.Log.Infof("CGO is not enabled! Cant dynamically load %s", soPlugsStr)
	pi.Log.Infof("Either enable CGO or remove RTPLUGS_SO from the environment")
	pi.Log.Infof("When CGO is not enabled use static loading of the desired plugs")
}
