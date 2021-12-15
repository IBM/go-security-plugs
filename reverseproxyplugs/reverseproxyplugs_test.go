package reverseproxyplugs

import (
	"net/http"
	"testing"
)

/*
type plug struct{}

func (p plug) Initialize(l pluginterfaces.Logger) {
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
*/
type countLog int

func (countLog) Debugf(format string, args ...interface{}) {
}
func (countLog) Infof(format string, args ...interface{}) {
}
func (c countLog) Warnf(format string, args ...interface{}) {
	c++
}
func (c countLog) Errorf(format string, args ...interface{}) {
	c++
}

func TestLoadPlugs(t *testing.T) {
	var config map[string]interface{}
	var numTests int

	if numTests = LoadPlugs(nil, "", nil); numTests != 0 {
		t.Errorf("LoadPlugs expected 0 returned %d\n", numTests)
	}

	if numTests = LoadPlugs(nil, "", nil); numTests != 0 {
		t.Errorf("LoadPlugs expected 0 returned %d\n", numTests)
	}

	if numTests = LoadPlugs(nil, "../plugs", nil); numTests != 1 {
		t.Errorf("LoadPlugs expected 1 returned %d\n", numTests)
	}

	var log countLog
	if numTests = LoadPlugs(log, "../plugs", config); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}
	if log > 0 {
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", log)
	}

}

//handleRequest
//HandleRequestPlugs
//HandleResponsePlugs
//HandleErrorPlugs
//ShutdownPlugs

func TestHandleRequestPlugs(t *testing.T) {
	type args struct {
		h http.Handler
	}

	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		// TODO: Add test cases.
		{"", args{nil}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandleRequestPlugs(tt.args.h); got != tt.want {
				t.Errorf("HandleRequestPlugs() returned = %v, want %v", got, tt.want)
			}
		})
	}
}
