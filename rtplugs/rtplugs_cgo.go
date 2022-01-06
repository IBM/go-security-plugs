//go:build cgo
// +build cgo

package rtplugs

import (
	"os"
	"plugin"
	"strings"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

// load() will attempt to dynamically load plugins
//
// env RTPLUGS_SO_PLUGINS defines a comma seperated list of .so plugin files
// relative/full path may be used
func load() {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic while loading so plugins. Recover: %v", r)
		}
	}()
	soPluginsStr := os.Getenv("RTPLUGS_SO")
	if soPluginsStr == "" {
		return
	}

	soPlugins := strings.Split(soPluginsStr, ",")
	pi.Log.Infof("Trying to load these %d plugins %v", len(soPlugins), soPlugins)

	for _, plugPkgPath := range soPlugins {
		_, err := plugin.Open(plugPkgPath)
		if err != nil {
			cwd, _ := os.Getwd()
			pi.Log.Infof("Plugin %s (cwd %s) dynamic loading skipped!: %v", plugPkgPath, cwd, err)
		} else {
			pi.Log.Infof("Plugin %s dynamic loading success!", plugPkgPath)
		}
	}
}
