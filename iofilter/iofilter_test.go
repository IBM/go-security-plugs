package iofilter

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func filterOk(buf []byte) error {
	fmt.Printf("filterOk: %d\n", len(buf))
	return nil
}

func filterErr(buf []byte) error {
	return errors.New("Not good")
}

func filterPanic(buf []byte) error {
	panic("OMG...")
}

type unothodoxReader struct {
	state      int
	closePanic bool
}

func (r *unothodoxReader) Read(buf []byte) (int, error) {
	if r.state == 0 {
		r.state = 1
		return 0, nil
	} else {
		r.state = 0
		return 0, errors.New("Aha!")
	}
}

func (r *unothodoxReader) Close() error {
	if r.closePanic {
		panic("I am already closed!")

	}
	return nil
}

func TestNewIoFilterBadReader(t *testing.T) {
	ur := new(unothodoxReader)

	t.Run("unothodoxReader", func(t *testing.T) {
		ur.closePanic = false
		r1 := NewIoFilter(ur, filterOk, 7)

		err1 := iotest.TestReader(r1, []byte(""))
		if err1 != nil {
			t.Fatal(err1)
		}

		if err := r1.Close(); err != nil {
			t.Errorf("iofilter.Close() error = %v", err)
		}

		ur.closePanic = true
		r2 := NewIoFilter(ur, filterOk, 7)
		err2 := iotest.TestReader(r2, []byte(""))
		if err2 != nil {
			t.Fatal(err2)
		}
		if err := r2.Close(); err != nil {
			t.Errorf("iofilter.Close() error = %v", err)
		}

	})
}

func TestNewIoFilter(t *testing.T) {
	const msg0 = ""
	const msg1 = "Now is the time for all good gophers."
	msg2Bytes := make([]byte, 81921)
	rand.Read(msg2Bytes[:])
	msg2 := string(msg2Bytes[:])
	tests := []struct {
		name    string
		size    int
		msg     string
		filter  func(buf []byte) error
		success bool
		panic   bool
	}{
		{"nofilter0", 0, msg0, nil, true, false},
		{"nofilter1", 0, msg1, nil, true, false},
		{"nofilter2", 0, msg2, nil, true, false},
		{"filterOk0", 0, msg0, filterOk, true, false},
		{"filterOk1", 0, msg1, filterOk, true, false},
		{"filterOk2", 0, msg2, filterOk, true, false},
		{"filterErr", 0, msg1, filterErr, false, false},
		{"filterPanic", 0, msg1, filterPanic, false, false},
		{"filterOk 1.1", 1, msg1, filterOk, true, false},
		{"filterOk 1.2", 2, msg1, filterOk, true, false},
		{"filterOk 1.3", 3, msg1, filterOk, true, false},
		{"filterOk 1.4", 4, msg1, filterOk, true, false},
		{"filterOk 1.8192", 8192, msg1, filterOk, true, false},
		{"filterOk 1.819200", 819200, msg1, filterOk, true, false},
		{"filterOk 2.3", 3, msg2, filterOk, true, false},
		{"filterOk 2.4", 4, msg2, filterOk, true, false},
		{"filterOk 2.8192", 8192, msg2, filterOk, true, false},
		{"filterOk 2.819200", 819200, msg2, filterOk, true, false},
		{"filterOk 1.-1", -1, msg1, filterOk, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.panic {
					fmt.Printf("r is %v \n", r)
					if r == nil {
						t.Errorf("The code did not panic")
					}
				} else {
					if r != nil {
						t.Errorf("The code paniced unexpectedly")
					}
				}
			}()
			var r io.ReadCloser
			if tt.filter == nil {
				r = io.NopCloser(strings.NewReader(tt.msg))
			} else {
				switch {
				case tt.size > 0:
					r = NewIoFilter(io.NopCloser(strings.NewReader(tt.msg)), tt.filter, tt.size)
				case tt.size < 0:
					r = NewIoFilter(io.NopCloser(strings.NewReader(tt.msg)), tt.filter, 1, 2)
				default:
					r = NewIoFilter(io.NopCloser(strings.NewReader(tt.msg)), tt.filter)
				}

			}
			err := iotest.TestReader(r, []byte(tt.msg))
			if (err == nil) != tt.success {
				t.Fatal(err)
			}

		})
	}
}
