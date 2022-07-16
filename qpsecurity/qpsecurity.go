package qpsecurity

import (
	"fmt"

	"github.com/IBM/go-security-plugs/rtplugs"
	"knative.dev/serving/pkg/queue/sharedmain"
)

type QPSecurityPlugs struct {
	rt *rtplugs.RoundTrip // list of activated plugs
}

func NewQPSecurityPlugs() *QPSecurityPlugs {
	fmt.Println("QPSecurityPlugs Init")
	return new(QPSecurityPlugs)
}

func (p *QPSecurityPlugs) QPTransport(qOpts *sharedmain.QPTransportOption) {
	p.rt = rtplugs.New(qOpts.Logger) // add qOpts.Context
	if p.rt != nil {
		qOpts.Logger.Infof("QPTransport setup plugs found!\n")
		qOpts.Transport = p.rt.Transport(qOpts.Transport)
	} else {
		qOpts.Logger.Infof("QPTransport setup no plugs found..\n")
	}
}

//func (p *QPSecurityPlugs) Init(ctx context.Context, logger *zap.SugaredLogger) {
//	p.logger = logger
//	p.rt = rtplugs.New(logger)
//	if p.rt == nil {
//		p.logger.Infof("QPSecurityPlugs Init - no plugs found\n")
//	}
//}

func (p *QPSecurityPlugs) Shutdown() {
	if p.rt != nil {
		p.rt.Close()
	}
}

// If extension does not require to be added to Transport
// (e.g. when the extensoin is not active),
// Transport should return next (never return nil)
//func (p *QPSecurityPlugs) Transport(next http.RoundTripper) (roundTripper http.RoundTripper) {
//	if p.rt == nil {
//		p.logger.Infof("QPSecurityPlugs Transport skipped\n")
//		return next
//	}
//	return p.rt.Transport(next)
//}
