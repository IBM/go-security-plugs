package rtplugs

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
var testconfig []string
var emptytestconfig []string
var falsetestconfig []string
var nokeyconfig []string
var panicconfig []string
var wtest http.ResponseWriter
var reqtest *http.Request

var resptest *http.Response
var errTest = "fake error"

var etest error
var defaultLog = dLog{}
var rt *RoundTrip

func init() {
	testconfig = []string{"../plugs/rtgate/rtgate.so",
		"../plugs/nopluggate/nopluggate.so",
		"../plugs/wrongpluggate/wrongpluggate.so",
		"../plugs/badversionplug/badversionplug.so",
	}
	//testconfig["panic"] = false
	emptytestconfig = []string{}
	falsetestconfig = []string{"path/to/nowhere"}
	panicconfig = []string{"../plugs/panicgate/panicgate.so"}

	reqtest, _ = http.NewRequest("GET", "http://10.0.0.1/", nil)
	//u, _ := url.Parse("http://1.2.3.4:5678")
	//reqtest.URL = u
	//reqtest.Header.Set("name", "value")
	resptest = &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString("Hello World")),
	}

	resptest.Request = reqtest
}
func InitializeEnv(panic string, errReq string, errResp string) {
	os.Setenv("RT_GATE_PANIC_INIT", "false")
	os.Setenv("RT_GATE_PANIC_SHUTDOWN", "false")
	os.Setenv("RT_GATE_PANIC_REQ", "false")
	os.Setenv("RT_GATE_PANIC_RESP", "false")
	os.Setenv("RT_GATE_PANIC_ERROR_REQ", "")
	os.Setenv("RT_GATE_PANIC_ERROR_RESP", "")
	if panic != "" {
		os.Setenv(panic, "true")
	}
	os.Setenv("RT_GATE_ERROR_REQ", errReq)
	os.Setenv("RT_GATE_ERROR_RESP", errResp)
}

type FakeRoundTrip struct{}

func (rt *FakeRoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	fmt.Println("Fake Round Trip!!!")
	return resptest, nil
}

func TestMain(m *testing.M) {
	testlog = 0
	InitializeEnv("", "", "")
	code := m.Run()
	os.Exit(code)
}

func TestUnloadPlugs(t *testing.T) {
	t.Run("", func(t *testing.T) {
		UnloadPlugs(nil)
		InitializeEnv("", "", "")
		rt = LoadPlugs(nil, testconfig)
		UnloadPlugs(rt)
		InitializeEnv("RT_GATE_PANIC_SHUTDOWN", "", "")
		rt = LoadPlugs(nil, testconfig)
		UnloadPlugs(rt)
	})
}

func TestTransport(t *testing.T) {
	var fake, roundtripper http.RoundTripper
	var rt *RoundTrip
	fake = new(FakeRoundTrip)

	t.Run("", func(t *testing.T) {
		var err error
		var resp *http.Response
		rt = LoadPlugs(nil, testconfig)
		roundtripper = Transport(rt, nil)
		roundtripper = Transport(rt, fake)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err != nil {
			t.Errorf("Transport returned with err %v\n", err)
		}
		if resp == nil {
			t.Errorf("Transport returned resp nil\n")
		}
		UnloadPlugs(rt)
		InitializeEnv("RT_GATE_PANIC_REQ", "", "")
		rt = LoadPlugs(nil, testconfig)
		roundtripper = Transport(rt, fake)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		UnloadPlugs(rt)
		InitializeEnv("RT_GATE_PANIC_RESP", "", "")
		rt = LoadPlugs(nil, testconfig)
		roundtripper = Transport(rt, fake)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		UnloadPlugs(rt)
		InitializeEnv("", "fake error", "")
		rt = LoadPlugs(nil, testconfig)
		roundtripper = Transport(rt, fake)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		UnloadPlugs(rt)
		InitializeEnv("", "", "fake error")
		rt = LoadPlugs(nil, testconfig)
		roundtripper = Transport(rt, fake)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		UnloadPlugs(rt)
	})
}

func TestLoadPlugs(t *testing.T) {
	var rt *RoundTrip

	if rt = LoadPlugs(nil, nil); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	t.Logf("emptytestconfig is %v", emptytestconfig)
	if rt = LoadPlugs(defaultLog, emptytestconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}
	t.Logf("falsetestconfig is %v", falsetestconfig)
	if rt = LoadPlugs(defaultLog, falsetestconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	if rt = LoadPlugs(defaultLog, nokeyconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	testlog = 0
	t.Logf("testconfig is %v", testconfig)
	if rt = LoadPlugs(nil, testconfig); rt == nil {
		t.Errorf("LoadPlugs did not expect nil\n")
	}

	if testlog > 0 {
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", testlog)
	}

	UnloadPlugs(rt)

	testlog = 0
	if rt = LoadPlugs(testlog, testconfig); rt == nil {
		t.Errorf("LoadPlugs did not expect nil\n")
	}
	if testlog > 0 {
		t.Errorf("LoadPlugs expected 0 warnings and errors, received  %d\n", testlog)
	}

	UnloadPlugs(rt)
	InitializeEnv("RT_GATE_PANIC_INIT", "", "")
	if rt = LoadPlugs(defaultLog, testconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

}

func TestDefaultLog(t *testing.T) {
	defaultLog.Debugf("Debugf")
	defaultLog.Infof("Infof")
	defaultLog.Warnf("Warnf")
	defaultLog.Errorf("Errorf")
}
