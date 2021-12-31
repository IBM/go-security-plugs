package rtplugs

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	goLog "log"
	"net/http"
	"os"
	"testing"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
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
var testconfigAll []string
var emptytestconfig []string
var falsetestconfig []string
var nokeyconfig []string
var panicconfig []string

//var wtest http.ResponseWriter
var reqtest *http.Request
var reqtestBlock *http.Request

var resptest *http.Response

//var errTest = "fake error"

//var etest error
type dLog struct{}

func (dLog) Debugf(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}
func (dLog) Infof(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}
func (dLog) Warnf(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}
func (dLog) Errorf(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}

var defaultLog = dLog{}
var rt *RoundTrip

func init() {
	testconfig = []string{"../plugs/rtgate/rtgate.so"}
	testconfigAll = []string{"../plugs/rtgate/rtgate.so",
		"../plugs/nopluggate/nopluggate.so",
		"../plugs/wrongpluggate/wrongpluggate.so",
		"../plugs/badversionplug/badversionplug.so",
	}
	//testconfig["panic"] = false
	emptytestconfig = []string{}
	falsetestconfig = []string{"path/to/nowhere"}
	panicconfig = []string{"../plugs/panicgate/panicgate.so"}

	reqtest, _ = http.NewRequest("GET", "http://10.0.0.1/", nil)
	reqtestBlock, _ = http.NewRequest("GET", "http://10.0.0.1/", nil)
	reqtestBlock.Header.Set("X-Block-Async", "0.01s")
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

type FakeRoundTrip struct {
}

var fakeRoundTripError = false

func (rt *FakeRoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	fmt.Println("Fake Round Trip!!! started")
	if fakeRoundTripError {
		fmt.Println("Fake Round Trip With Error!!!")
		return resptest, errors.New("fake error")
	}
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
		InitializeEnv("", "", "")
		rt = New(testconfig)
		rt.Close()
		InitializeEnv("RT_GATE_PANIC_SHUTDOWN", "", "")
		rt = New(testconfig)
		rt.Close()
	})
}

func TestTransport(t *testing.T) {
	var fakeRoundTrip, roundtripper http.RoundTripper
	var rt *RoundTrip
	fakeRoundTrip = new(FakeRoundTrip)
	_ = fakeRoundTrip
	t.Run("", func(t *testing.T) {
		var err error
		var resp *http.Response

		// Async Timeout with default transport
		InitializeEnv("", "", "")
		rt = New(testconfig)
		roundtripper = rt.Transport(nil)
		resp, err = roundtripper.RoundTrip(reqtestBlock)
		fmt.Printf("TestTransport 4\n")
		if err == nil {
			t.Errorf("Transport returned without err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport returned resp not nil\n")
		}
		rt.Close()

		// Fake transport
		fakeRoundTripError = false
		InitializeEnv("", "", "")
		rt = New(testconfig)
		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err != nil {
			t.Errorf("Transport returned with err %v\n", err)
		}
		if resp == nil {
			t.Errorf("Transport returned resp nil\n")
		}
		rt.Close()

		// Fake transport with error
		fakeRoundTripError = true
		InitializeEnv("", "", "")
		rt = New(testconfig)
		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		rt.Close()
		fakeRoundTripError = false

		// Fake transport with RTGate Panic at REQ
		InitializeEnv("RT_GATE_PANIC_REQ", "", "")
		rt = New(testconfig)

		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		rt.Close()

		// Fake transport with RTGate Panic at Resp
		InitializeEnv("RT_GATE_PANIC_RESP", "", "")
		rt = New(testconfig)
		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		rt.Close()

		// Fake transport with RTGate Error at Req
		InitializeEnv("", "fake error", "")
		rt = New(testconfig)
		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		rt.Close()

		// Fake transport with RTGate Error at Resp
		InitializeEnv("", "", "fake error")
		rt = New(testconfig)
		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		rt.Close()
	})
}

func TestLoadPlugs(t *testing.T) {
	var rt *RoundTrip

	if rt = New(nil); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	t.Logf("emptytestconfig is %v", emptytestconfig)
	if rt = New(emptytestconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	t.Logf("falsetestconfig is %v", falsetestconfig)
	if rt = New(falsetestconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	if rt = New(nokeyconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	t.Logf("testconfig is %v", testconfigAll)
	if rt = New(testconfigAll); rt == nil {
		t.Errorf("LoadPlugs did not expect nil\n")
	}
	rt.Close()

	log := pi.Log
	pi.Log = testlog
	rt = New(testconfig)
	rt.Close()

	pi.Log = log

	InitializeEnv("RT_GATE_PANIC_INIT", "", "")
	if rt = New(testconfig); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}
}

func TestDefaultLog(t *testing.T) {
	defaultLog.Debugf("Debugf")
	defaultLog.Infof("Infof")
	defaultLog.Warnf("Warnf")
	defaultLog.Errorf("Errorf")
}
