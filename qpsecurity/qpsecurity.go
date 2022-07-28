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
	p.defaults.Logger.Infof("QPSecurityPlugs ProcessAnnotations started")
	file, err := os.Open("/etc/podinfo/annotations")
	if err != nil {
		p.defaults.Logger.Infof("QPSecurityPlugs failed to open /etc/podinfo/annotations. Check if podInfo is enabled for this service. os.Open Error %s", err.Error())
		return
	}
	defer file.Close()
	p.config = make(map[string]map[string]string)
	p.plugs = make([]string, 0)

	qpextentionPreifx := "qpextention.knative.dev/"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := scanner.Text()
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
		p.defaults.Logger.Infof("QPSecurityPlugs scanner Error %s", err.Error())
		return
	}
	p.defaults.Logger.Infof("QPSecurityPlugs activated plugs: %v", p.plugs)
	p.defaults.Logger.Infof("QPSecurityPlugs config plugs: %v", p.config)
}

func (p *QPSecurityPlugs) Setup(defaults *sharedmain.Defaults) {
	p.defaults = defaults
	p.defaults.Logger.Infof("QPSecurityPlugs Setup started")
	p.ProcessAnnotations()
	p.rt = rtplugs.NewConfigrablePlugs(p.plugs, p.config, defaults.Logger) // add qOpts.Context
	if p.rt != nil {
		defaults.Transport = p.rt.Transport(defaults.Transport)
	} else {
		defaults.Logger.Infof("Setup no active plugs found...")
	}
}

func (p *QPSecurityPlugs) Shutdown() {
	if p.rt != nil {
		p.rt.Close()
	}
}
