package qpsecurity

import (
	"context"
	"net/http"

	"github.com/IBM/go-security-plugs/rtplugs"
	"go.uber.org/zap"
)

type QPSecurityPlugs struct {
	rt     *rtplugs.RoundTrip // list of activated plugs
	logger *zap.SugaredLogger
}

var SecurityExtensions QPSecurityPlugs

func (p *QPSecurityPlugs) Init(logger *zap.SugaredLogger, ctx context.Context) {
	p.logger = logger
	p.rt = rtplugs.New(logger)
	if p.rt == nil {
		p.logger.Infof("QPSecurityPlugs Init - no plugs found\n")
	}
}

func (p *QPSecurityPlugs) Shutdown() {
	if p.rt != nil {
		p.rt.Close()
	}
}

// If extension does not require to be added to Transport
// (e.g. when the extensoin is not active),
// Transport should return next (never return nil)
func (p *QPSecurityPlugs) Transport(next http.RoundTripper) (roundTripper http.RoundTripper) {
	if p.rt == nil {
		p.logger.Infof("QPSecurityPlugs Transport skipped\n")
		return next
	}
	return p.rt.Transport(next)
}
