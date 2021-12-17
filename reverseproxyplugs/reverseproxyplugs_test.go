package reverseproxyplugs

import (
	"errors"
	"net/http"
	"os"
	"testing"
)

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

var testlog countLog
var testconfig map[string]interface{} = make(map[string]interface{})
var emptytestconfig map[string]interface{} = make(map[string]interface{})
var falsetestconfig map[string]interface{} = make(map[string]interface{})
var nokeyconfig map[string]interface{} = make(map[string]interface{})
var wtest http.ResponseWriter
var reqtest http.Request
var resptest http.Response
var errTest = errors.New("fake error")

var etest error

func init() {
	testconfig["reverseproxyplugins"] = []string{"../plugs/examplegate/examplegate.so",
		"../plugs/panicgate/panicgate.so",
		"../plugs/nopluggate/nopluggate.so",
		"../plugs/wrongpluggate/wrongpluggate.so",
		"../plugs/badversionplug/badversionplug.so",
	}
	testconfig["panic"] = false
	emptytestconfig["reverseproxyplugins"] = []string{}
	falsetestconfig["reverseproxyplugins"] = []int{2}
}

func TestMain(m *testing.M) {
	testconfig["panic"] = false
	testconfig["error"] = nil
	LoadPlugs(defaultLog, testconfig)
	code := m.Run()
	UnloadPlugs()
	testlog = 0
	os.Exit(code)
}
func TestLoadPlugs(t *testing.T) {
	var numTests int
	if numTests = LoadPlugs(testlog, testconfig); numTests != 4 {
		t.Errorf("LoadPlugs expected 24returned %d\n", numTests)
	}
	t.Errorf("LoadPlugs")
	return

	if numTests = LoadPlugs(nil, nil); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}

	if numTests = LoadPlugs(defaultLog, emptytestconfig); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}

	if numTests = LoadPlugs(defaultLog, falsetestconfig); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}

	if numTests = LoadPlugs(defaultLog, nokeyconfig); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}

	testlog = 0
	if numTests = LoadPlugs(testlog, testconfig); numTests != 4 {
		t.Errorf("LoadPlugs expected 24returned %d\n", numTests)
	}
	if testlog > 0 {
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", log)
	}

	UnloadPlugs()
	testlog = 0
	if numTests = LoadPlugs(testlog, testconfig); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}
	if testlog > 0 {
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", log)
	}

	testconfig["panic"] = true
	if numTests = LoadPlugs(defaultLog, testconfig); numTests != 3 {
		t.Errorf("LoadPlugs expected 3 returned %d\n", numTests)
	}
	testconfig["panic"] = false
}

//handleRequest
//HandleRequestPlugs
//HandleResponsePlugs
//HandleErrorPlugs
//ShutdownPlugs

func TestDefaultLog(t *testing.T) {
	defaultLog.Debugf("Debugf")
	defaultLog.Infof("Infof")
	defaultLog.Warnf("Warnf")
	defaultLog.Errorf("Errorf")
}

func TestHandleRequestPlugs(t *testing.T) {
	var success bool = false

	finalh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		success = true
	})
	type args struct {
		h http.Handler
	}

	tests := []struct {
		name string
		args args
		err  error
	}{
		// TODO: Add test cases.
		{"", args{finalh}, nil},
		{"", args{finalh}, errTest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testconfig["error"] = tt.err
			got := HandleRequestPlugs(tt.args.h)
			if got == nil {
				t.Errorf("HandleRequestPlugs() returned = %v which is unwated", got)
			}

			got.ServeHTTP(wtest, &reqtest)
			if (tt.err == nil) != success {
				t.Errorf("HandleRequestPlugs - ServeHTTP failed")
			}
			success = false
			testconfig["panic"] = true
			got.ServeHTTP(wtest, &reqtest)
			if success {
				t.Errorf("HandleRequestPlugs - ServeHTTP succeded when we expected it to panic and fail")
			}
			success = false
			testconfig["panic"] = false
			testconfig["error"] = nil
		})
	}
}

func TestHandleErrorPlugs(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
		e error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"", args{wtest, &reqtest, etest}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleErrorPlugs(tt.args.w, tt.args.r, tt.args.e)
			testconfig["panic"] = true
			HandleErrorPlugs(tt.args.w, tt.args.r, tt.args.e)
			testconfig["panic"] = false
			testconfig["error"] = errTest
			HandleErrorPlugs(tt.args.w, tt.args.r, tt.args.e)
			testconfig["error"] = nil
		})
	}
}

func TestHandleResponsePlugs(t *testing.T) {
	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"Without Error", args{&resptest}, false},
		{"With Error", args{&resptest}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				testconfig["error"] = errTest
			}
			if err := HandleResponsePlugs(tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("HandleResponsePlugs() error = %v, wantErr %v", err, tt.wantErr)
			}
			testconfig["panic"] = true
			if err := HandleResponsePlugs(tt.args.resp); err != nil {
				t.Errorf("HandleResponsePlugs() error = %v, wantErr %v", err, tt.wantErr)
			}
			testconfig["panic"] = false
			testconfig["error"] = nil
		})
	}
}

func TestUnloadPlugs(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UnloadPlugs()
			LoadPlugs(defaultLog, testconfig)
			testconfig["panic"] = true
			UnloadPlugs()
			testconfig["panic"] = false
		})
	}
}
