package testgate

import (
	"net/http"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"
const name string = "testgate"

type plug struct {
	name    string
	version string

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
		pi.Log.Infof("%s: hehe, someone noticed me!", p.name)
	}
	return req, nil
}

func (p *plug) ApproveResponse(req *http.Request, resp *http.Response) (*http.Response, error) {
	if _, ok := req.Header["X-Testgate-Hi"]; ok {
		resp.Header.Add("X-Testgate-Bye", "CU")
	}
	return resp, nil
}

func (p *plug) Shutdown() {
	pi.Log.Infof("%s: Shutdown", p.name)
}

func (p *plug) Init() {
	pi.Log.Infof("plug %s: Initializing - version %v", p.name, p.version)
	pi.Log.Infof("plug %s: Never use in production", p.name)
}

func init() {
	p := new(plug)
	p.version = version
	p.name = name
	pi.RegisterPlug(p)
}
