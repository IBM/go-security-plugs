package reverseproxyplugs

import (
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
func (c countLog) Sync() error {
	return nil
}

var testlog countLog
var testconfig []string
var emptytestconfig []string
var falsetestconfig []string
var nokeyconfig []string
var wtest http.ResponseWriter
var reqtest http.Request
var resptest http.Response
var errTest = "fake error"

var etest error
var defaultLog = dLog{}

func init() {
	testconfig = []string{"../plugs/examplegate/examplegate.so",
		"../plugs/panicgate/panicgate.so",
		"../plugs/nopluggate/nopluggate.so",
		"../plugs/wrongpluggate/wrongpluggate.so",
		"../plugs/badversionplug/badversionplug.so",
	}
	//testconfig["panic"] = false
	emptytestconfig = []string{}
	falsetestconfig = []string{"path/to/nowhere"}
	resptest.Request = &reqtest
}
func InitializeEnv(panic string, err string) {
	UnloadPlugs()
	os.Setenv("PANIC_GATE_PANIC_INIT", "false")
	os.Setenv("PANIC_GATE_PANIC_SHUTDOWN", "false")
	os.Setenv("PANIC_GATE_PANIC_REQ", "false")
	os.Setenv("PANIC_GATE_PANIC_RESP", "false")
	os.Setenv("PANIC_GATE_PANIC_ERR", "false")
	if panic != "" {
		os.Setenv(panic, "true")
	}
	os.Setenv("PANIC_GATE_ERROR", err)
	LoadPlugs(defaultLog, testconfig)
}

func TestMain(m *testing.M) {
	testlog = 0
	UnloadPlugs()
	code := m.Run()
	os.Exit(code)
}
func TestLoadPlugs(t *testing.T) {
	var numTests int

	InitializeEnv("", "")

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
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", testlog)
	}

	UnloadPlugs()
	testlog = 0
	if numTests = LoadPlugs(testlog, testconfig); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}
	if testlog > 0 {
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", testlog)
	}

	InitializeEnv("PANIC_GATE_PANIC_INIT", "")
	if numTests = LoadPlugs(defaultLog, testconfig); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}
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
		name  string
		args  args
		err   string
		panic bool
	}{
		// TODO: Add test cases.
		{"", args{finalh}, "", false},
		{"", args{finalh}, errTest, false},
		{"", args{finalh}, "", true},
		{"", args{finalh}, errTest, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				InitializeEnv("PANIC_GATE_PANIC_REQ", "true")
			} else {
				InitializeEnv("", tt.err)
			}
			got := HandleRequestPlugs(tt.args.h)
			if got == nil {
				t.Errorf("HandleRequestPlugs() returned = %v which is unwated", got)
			}

			success = false
			got.ServeHTTP(wtest, &reqtest)
			if tt.panic && success {
				t.Errorf("HandleRequestPlugs - ServeHTTP succeded when we expected it to panic and fail")
			}
			if (tt.err != "") && success {
				t.Errorf("HandleRequestPlugs - ServeHTTP succeded when we expected it to error and fail")
			}
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
		name  string
		args  args
		panic bool
	}{
		// TODO: Add test cases.
		{"", args{wtest, &reqtest, etest}, false},
		{"", args{wtest, &reqtest, etest}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				InitializeEnv("PANIC_GATE_PANIC_REQ", "true")
			} else {
				InitializeEnv("", "")
			}
			HandleErrorPlugs(tt.args.w, tt.args.r, tt.args.e)
		})
	}
}

func TestHandleResponsePlugs(t *testing.T) {
	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name  string
		args  args
		err   string
		panic bool
	}{
		// TODO: Add test cases.
		{"Without Error", args{&resptest}, "", false},
		{"With Error", args{&resptest}, errTest, false},
		{"Panic Without Error", args{&resptest}, "", true},
		{"Panic With Error", args{&resptest}, errTest, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				InitializeEnv("PANIC_GATE_PANIC_RESP", "true")
			} else {
				InitializeEnv("", tt.err)
			}

			err := HandleResponsePlugs(tt.args.resp)
			if tt.panic {
				if err.Error() != "plug paniced" {
					t.Errorf("HandleResponsePlugs() panic but err is %v", err)
				}
			} else {
				if (err == nil) == (tt.err != "") {
					t.Errorf("HandleResponsePlugs() error = %v, expected %v", err, tt.err)
				}
			}
		})
	}
}

func TestUnloadPlugs(t *testing.T) {
	t.Run("", func(t *testing.T) {
		InitializeEnv("", "")
		UnloadPlugs()
		InitializeEnv("PANIC_GATE_PANIC_SHUTDOWN", "")
		UnloadPlugs()
	})

}
