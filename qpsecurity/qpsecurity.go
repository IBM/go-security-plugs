package qpsecurity

import (
	"github.com/IBM/go-security-plugs/rtplugs"
	"knative.dev/serving/pkg/queue/sharedmain"
)

type QPSecurityPlugs struct {
	rt *rtplugs.RoundTrip // list of activated plugs
}

func NewQPSecurityPlugs() *QPSecurityPlugs {
	return new(QPSecurityPlugs)
}

func (p *QPSecurityPlugs) Setup(d *sharedmain.Defaults) {
	p.rt = rtplugs.New(d.Logger) // add qOpts.Context
	if p.rt != nil {
		d.Transport = p.rt.Transport(d.Transport)
	} else {
		d.Logger.Infof("Setup no active plugs found...")
	}
}

func (p *QPSecurityPlugs) Shutdown() {
	if p.rt != nil {
		p.rt.Close()
	}
}
