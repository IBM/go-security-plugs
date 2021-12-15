package reverseproxyplugs

import (
	"net/http"
	"testing"
)

type plug struct{}

func (p plug) Initialize(l Logger) {
	log = l
	log.Infof("OK")
}

func (plug) Shutdown() {
	log.Infof("OK")
}

func (plug) PlugName() string {
	return "OK"
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	log.Infof("OK")
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	log.Infof("OK")
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	log.Infof("OK")
	return nil
}

var Plug plug

func TestLoadPlugs(t *testing.T) {
	var ext []string
	var numTests int
	if numTests = LoadPlugs(nil, nil); numTests != 0 {
		t.Errorf("LoadPlugs expected 0 returned %d\n", numTests)
	}

	ext = []string{}
	if numTests = LoadPlugs(nil, ext); numTests != 0 {
		t.Errorf("LoadPlugs expected 0 returned %d\n", numTests)
	}
}

//handleRequest
//HandleRequestPlugs
//HandleResponsePlugs
//HandleErrorPlugs
//ShutdownPlugs
