package main

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/IBM/go-security-plugs/historical/pluginterfaces"
)

type dLog struct {
}

func (d dLog) Debugf(format string, args ...interface{}) {}
func (d dLog) Infof(format string, args ...interface{})  {}
func (d dLog) Warnf(format string, args ...interface{})  {}
func (d dLog) Errorf(format string, args ...interface{}) {}
func (d dLog) Sync() error                               { return nil }

var defaultLog dLog
var p plug

func TestMain(m *testing.M) {
	p.Initialize()
	code := m.Run()
	os.Exit(code)
}

func Test_plug_Initialize(t *testing.T) {
	type args struct {
		l pluginterfaces.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"Log args", args{defaultLog}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.Initialize()
		})
	}
}

func Test_plug_Shutdown(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.Shutdown()
		})
	}
}

func Test_plug_PlugName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{"", "ExampleGate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.PlugName(); got != tt.want {
				t.Errorf("plug.PlugName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_plug_ResponseHook(t *testing.T) {
	tests := []struct {
		name  string
		block bool
	}{
		// TODO: Add test cases.
		{"", false},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/some/path", nil)
			if tt.block {
				req.Header.Set("X-Block-Resp", "value")
			}
			respRecorder := httptest.NewRecorder()
			fmt.Fprintf(respRecorder, "Hi there!")
			resp := respRecorder.Result()
			resp.Request = req
			resp.Header.Set("name", "val")
			if err := p.ResponseHook(resp); (err != nil) != tt.block {
				t.Errorf("plug.ResponseHook() error = %v, wantErr %v", err, tt.block)
			}
		})
	}
}

func Test_plug_RequestHook(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		block      bool
	}{
		// TODO: Add test cases.
		{"", 200, false},
		{"", 200, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p plug
			p.Initialize()
			req := httptest.NewRequest("GET", "/some/path", nil)
			req.Header.Set("name", "value")
			if tt.block {
				req.Header.Set("X-Block-Req", "value")
			}
			resp := httptest.NewRecorder()

			if err := p.RequestHook(resp, req); (err != nil) != tt.block {
				t.Errorf("plug.RequestHook() error = %v, block %v", err, tt.block)
			}
			if resp.Code != tt.statusCode {
				t.Errorf("Want status '%d', got '%d'", tt.statusCode, resp.Code)
			}

			//if strings.TrimSpace(resp.Body.String()) != tt.want {
			//	t.Errorf("Want '%s', got '%s'", tt.want, resp.Body)
			//}
		})
	}
}

func Test_plug_PlugVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{"", version},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.PlugVersion(); got != tt.want {
				t.Errorf("plug.PlugVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_plug_ErrorHook(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/some/path", nil)
			resp := httptest.NewRecorder()
			e := errors.New("TestError")
			p.ErrorHook(resp, req, e)
		})
	}
}
