package testgate

import (
	"context"
	"net/http"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"
const name string = "testgate"

type plug struct {
	name    string
	version string
	config  map[string]string

	sender string
	answer string

	// Add here any other state the extension needs
}

func (p *plug) PlugName() string {
	return p.name
}

func (p *plug) PlugVersion() string {
	return p.version
}

func (p *plug) ApproveRequest(req *http.Request) (*http.Request, error) {
	if _, ok := req.Header["X-Testgate-Hi"]; ok {
		pi.Log.Infof("%s: hehe, %s noticed me!", p.name, p.sender)
	}
	return req, nil
}

func (p *plug) ApproveResponse(req *http.Request, resp *http.Response) (*http.Response, error) {
	if _, ok := req.Header["X-Testgate-Hi"]; ok {
		resp.Header.Add("X-Testgate-Bye", p.answer)
	}
	return resp, nil
}

func (p *plug) Shutdown() {
	pi.Log.Infof("%s: Shutdown", p.name)
}

func (p *plug) Start(ctx context.Context) context.Context {
	return ctx
}

func (p *plug) Init(c map[string]string) {
	p.config = c

	pi.Log.Infof("plug %s: Initializing - version %v", p.name, p.version)
	pi.Log.Infof("plug %s: Never use in production", p.name)
	pi.Log.Infof("plug %s: has config %v", p.name, p.config)
	p.answer = "CU"
	p.sender = "someone"
	if p.config != nil {
		if v, ok := p.config["sender"]; ok {
			p.sender = v
			pi.Log.Infof("plug %s: found sender %s", p.name, p.sender)
		}
		if v, ok := p.config["response"]; ok {
			p.answer = v
			pi.Log.Infof("plug %s: found answer %s", p.name, p.answer)
		}
	}
}

func init() {
	p := new(plug)
	p.version = version
	p.name = name
	pi.RegisterPlug(p)
}
