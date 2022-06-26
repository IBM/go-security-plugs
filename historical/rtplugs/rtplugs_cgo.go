//go:build cgo
// +build cgo

package rtplugs

import (
	"os"
	"plugin"
	"strings"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

// load() will attempt to dynamically load plugs
//
// env RTPLUGS_SO defines a comma seperated list of .so plug files
// relative/full path may be used
func load() {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic while loading .so plugs. Recover: %v", r)
		}
	}()
	soPlugsStr := os.Getenv("RTPLUGS_SO")
	if soPlugsStr == "" {
		return
	}

	soPlugs := strings.Split(soPlugsStr, ",")
	pi.Log.Infof("Trying to load these %d plugs %v", len(soPlugs), soPlugs)

	for _, plugPkgPath := range soPlugs {
		_, err := plugin.Open(plugPkgPath)
		if err != nil {
			cwd, _ := os.Getwd()
			pi.Log.Infof("Plugin %s (cwd %s) dynamic loading skipped!: %v", plugPkgPath, cwd, err)
		} else {
			pi.Log.Infof("Plugin %s dynamic loading success!", plugPkgPath)
		}
	}
}
