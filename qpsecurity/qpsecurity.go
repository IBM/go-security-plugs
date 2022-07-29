package qpsecurity

import (
	"bufio"
	"os"
	"strings"

	"github.com/IBM/go-security-plugs/rtplugs"
	"knative.dev/serving/pkg/queue/sharedmain"
)

type QPSecurityPlugs struct {
	rt       *rtplugs.RoundTrip // list of activated plugs
	config   map[string]map[string]string
	defaults *sharedmain.Defaults
	plugs    []string
}

func NewQPSecurityPlugs() *QPSecurityPlugs {
	return new(QPSecurityPlugs)
}

func (p *QPSecurityPlugs) ProcessAnnotations() {
	file, err := os.Open("/etc/podinfo/annotations")
	if err != nil {
		p.defaults.Logger.Infof("Failed to open /etc/podinfo/annotations. Check if podInfo is mounted. os.Open Error %s", err.Error())
		return
	}
	defer file.Close()
	p.config = make(map[string]map[string]string)
	p.plugs = make([]string, 0)

	qpextentionPreifx := "qpextention.knative.dev/"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := scanner.Text()
		txt = strings.ToLower(txt)
		parts := strings.Split(txt, "=")

		k := parts[0]
		v := parts[1]
		if strings.HasPrefix(k, qpextentionPreifx) && len(k) > len(qpextentionPreifx) {
			v = strings.TrimSuffix(strings.TrimPrefix(v, "\""), "\"")
			k = k[len(qpextentionPreifx):]
			keyparts := strings.Split(k, "-")
			if len(keyparts) < 2 {
				continue
			}
			extension := keyparts[0]
			action := keyparts[1]
			switch action {
			case "activate":
				if !strings.EqualFold(v, "enable") {
					continue
				}
				p.plugs = append(p.plugs, extension)
			case "config":
				if len(keyparts) == 3 {
					extensionKey := keyparts[2]
					if _, ok := p.config[extension]; !ok {
						p.config[extension] = make(map[string]string)
					}
					p.config[extension][extensionKey] = v
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		p.defaults.Logger.Infof("Scanner Error %s", err.Error())
		return
	}
	p.defaults.Logger.Debugf("Plug: %v was activated with config %v", p.plugs, p.config)
}

func (p *QPSecurityPlugs) Setup(defaults *sharedmain.Defaults) {
	p.defaults = defaults
	servicename := defaults.Env.ServingService
	if servicename == "" {
		servicename = defaults.Env.ServingConfiguration
	}

	// build p.config
	p.ProcessAnnotations()

	p.rt = rtplugs.NewConfigrablePlugs(defaults.Ctx, defaults.Logger, servicename, defaults.Env.ServingNamespace, p.plugs, p.config) // add qOpts.Context
	if p.rt != nil {
		defaults.Ctx = p.rt.Start(defaults.Ctx)
		defaults.Transport = p.rt.Transport(defaults.Transport)
	} else {
		defaults.Logger.Infof("No plugs were activated")
	}
}

func (p *QPSecurityPlugs) Shutdown() {
	if p.rt != nil {
		p.rt.Close()
	}
}
