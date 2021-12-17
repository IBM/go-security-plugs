package main

import (
	"net/http"
	"testing"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

type dLog struct {
}

func (d dLog) Debugf(format string, args ...interface{}) {}
func (d dLog) Infof(format string, args ...interface{})  {}
func (d dLog) Warnf(format string, args ...interface{})  {}
func (d dLog) Errorf(format string, args ...interface{}) {}

var defaultLog dLog
var testconfig map[string]interface{}

func Test_plug_Initialize(t *testing.T) {
	type fields struct {
		version string
		log     pluginterfaces.Logger
	}
	type args struct {
		l pluginterfaces.Logger
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"Log args", fields{"myVer", nil}, args{defaultLog}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plug{
				version: tt.fields.version,
				log:     tt.fields.log,
			}
			p.Initialize(tt.args.l, testconfig)
		})
	}
}

func Test_plug_Shutdown(t *testing.T) {
	type fields struct {
		version string
		log     pluginterfaces.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{"", fields{"myVer", defaultLog}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plug{
				version: tt.fields.version,
				log:     tt.fields.log,
			}
			p.Shutdown()
		})
	}
}

func Test_plug_PlugName(t *testing.T) {
	type fields struct {
		version string
		log     pluginterfaces.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{"", fields{"myVer", defaultLog}, "ExampleGate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plug{
				version: tt.fields.version,
				log:     tt.fields.log,
			}
			if got := p.PlugName(); got != tt.want {
				t.Errorf("plug.PlugName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_plug_ResponseHook(t *testing.T) {
	type fields struct {
		version string
		log     pluginterfaces.Logger
	}
	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plug{
				version: tt.fields.version,
				log:     tt.fields.log,
			}
			if err := p.ResponseHook(tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("plug.ResponseHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_plug_RequestHook(t *testing.T) {
	type fields struct {
		version string
		log     pluginterfaces.Logger
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plug{
				version: tt.fields.version,
				log:     tt.fields.log,
			}
			if err := p.RequestHook(tt.args.w, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("plug.RequestHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
