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

	_ "github.com/IBM/go-security-plugs/plugs/rtgate"
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
var testconfig string
var testconfigAll string
var emptytestconfig string
var falsetestconfig string

// var wtest http.ResponseWriter
var reqtest *http.Request
var reqtestBlock *http.Request

var resptest *http.Response

//var errTest = "fake error"

// var etest error
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
	testconfig = "rtgate"
	testconfigAll = "noplug,rtgate,nothing"
	//testconfig["panic"] = false
	emptytestconfig = ""
	falsetestconfig = "noplug"

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
func InitializeEnv(params ...string) {
	fmt.Printf("InitializeEnv %d %v\n", len(params), params)

	switch len(params) {
	case 4:
		os.Setenv("RTPLUGS", params[3])
	default:
		os.Setenv("RTPLUGS", testconfig)
	}

	os.Setenv("NAMESPACE", "myns")
	os.Setenv("SERVICENAME", "myid")
	os.Setenv("RT_GATE_PANIC_INIT", "false")
	os.Setenv("RT_GATE_PANIC_SHUTDOWN", "false")
	os.Setenv("RT_GATE_PANIC_REQ", "false")
	os.Setenv("RT_GATE_PANIC_RESP", "false")
	os.Setenv("RT_GATE_PANIC_ERROR_REQ", "")
	os.Setenv("RT_GATE_PANIC_ERROR_RESP", "")
	os.Setenv("RT_GATE_ERROR_REQ", "")
	os.Setenv("RT_GATE_ERROR_RESP", "")
	if len(params) > 0 && params[0] != "" {
		os.Setenv(params[0], "true")
	}
	if len(params) > 1 && params[1] != "" {
		os.Setenv("RT_GATE_ERROR_REQ", params[1])
	}
	if len(params) > 2 && params[2] != "" {
		os.Setenv("RT_GATE_ERROR_RESP", params[2])
	}
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
		InitializeEnv()
		rt = New(nil)
		rt.Close()
		InitializeEnv("RT_GATE_PANIC_SHUTDOWN", "", "")
		rt = New(nil)
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
		InitializeEnv()
		rt = New(nil)
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
		InitializeEnv()
		rt = New(nil)
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
		InitializeEnv()
		rt = New(nil)
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
		InitializeEnv("RT_GATE_PANIC_REQ")
		rt = New(nil)

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
		InitializeEnv("RT_GATE_PANIC_RESP")
		rt = New(nil)
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
		InitializeEnv("", "fake error")
		rt = New(nil)
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
		rt = New(nil)
		roundtripper = rt.Transport(fakeRoundTrip)
		resp, err = roundtripper.RoundTrip(reqtest)
		if err == nil {
			t.Errorf("Transport should return err %v\n", err)
		}
		if resp != nil {
			t.Errorf("Transport should not return resp\n")
		}
		//t.Fatalf("STOP!\n")
		rt.Close()
	})
}

func TestLoadPlugs(t *testing.T) {
	var rt *RoundTrip

	InitializeEnv("", "", "", emptytestconfig)
	t.Logf("emptytestconfig is %v", emptytestconfig)
	if rt = New(nil); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	InitializeEnv("", "", "", falsetestconfig)
	t.Logf("falsetestconfig is %v", falsetestconfig)
	if rt = New(nil); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}

	InitializeEnv("", "", "", testconfig)
	t.Logf("testconfig is %v", testconfig)
	if rt = New(nil); rt == nil {
		t.Errorf("LoadPlugs  did not expect nil\n")
	}

	InitializeEnv("", "", "", testconfigAll)
	t.Logf("testconfig is %v", testconfigAll)
	if rt = New(nil); rt == nil {
		t.Errorf("LoadPlugs did not expect nil\n")
	}

	rt.Close()

	InitializeEnv()
	log := pi.Log
	rt = New(testlog)
	rt.Close()

	pi.Log = log

	InitializeEnv("RT_GATE_PANIC_INIT")
	if rt = New(nil); rt != nil {
		t.Errorf("LoadPlugs expected nil\n")
	}
}

func TestDefaultLog(t *testing.T) {
	defaultLog.Debugf("Debugf")
	defaultLog.Infof("Infof")
	defaultLog.Warnf("Warnf")
	defaultLog.Errorf("Errorf")
}
