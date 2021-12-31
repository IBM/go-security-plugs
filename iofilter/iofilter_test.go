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

func TestNewBadReader(t *testing.T) {
	ur := new(unothodoxReader)

	t.Run("unothodoxReader", func(t *testing.T) {
		ur.closePanic = false
		r1 := New(ur, filterOk, 7)

		err1 := iotest.TestReader(r1, []byte(""))
		if err1 != nil {
			t.Fatal(err1)
		}

		if err := r1.Close(); err != nil {
			t.Errorf("iofilter.Close() error = %v", err)
		}

		ur.closePanic = true
		r2 := New(ur, filterOk, 7)
		err2 := iotest.TestReader(r2, []byte(""))
		if err2 != nil {
			t.Fatal(err2)
		}
		if err := r2.Close(); err != nil {
			t.Errorf("iofilter.Close() error = %v", err)
		}

	})
}

func TestNew(t *testing.T) {
	const msg0 = ""
	const msg1 = "Now is the time for all good gophers."
	msg2Bytes := make([]byte, 256)
	rand.Read(msg2Bytes[:])

	msg2 := string(msg2Bytes[:])

	msgs := []string{msg0, msg1, msg2}

	numBufs := []uint{0, 1, 2, 3, 4, 8192}
	sizeBufs := []uint{0, 1, 2, 3, 4, 8192}

	for _, msg := range msgs {
		t.Run("", func(t *testing.T) {
			r := io.NopCloser(strings.NewReader(msg))
			err := iotest.TestReader(r, []byte(msg))
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("", func(t *testing.T) {
			r := New(io.NopCloser(strings.NewReader(msg)), filterOk)
			err := iotest.TestReader(r, []byte(msg))
			if err != nil {
				t.Fatal(err)
			}
		})
		for _, numBuf := range numBufs {
			t.Run("", func(t *testing.T) {
				r := New(io.NopCloser(strings.NewReader(msg)), filterOk, numBuf)
				err := iotest.TestReader(r, []byte(msg))
				if err != nil {
					t.Fatal(err)
				}
			})
			for _, sizeBuf := range sizeBufs {
				t.Run("", func(t *testing.T) {
					r := New(io.NopCloser(strings.NewReader(msg)), filterOk, numBuf, sizeBuf)
					err := iotest.TestReader(r, []byte(msg))
					if err != nil {
						t.Fatal(err)
					}
				})
			}
		}

	}

	t.Run("", func(t *testing.T) {
		defer func() {
			r := recover()
			fmt.Printf("r is %v \n", r)
			if r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		r := New(io.NopCloser(strings.NewReader(msg1)), filterOk, 1, 2, 3)
		err := iotest.TestReader(r, []byte(msg1))
		if err == nil {
			t.Fatal(err)
		}
	})
	t.Run("", func(t *testing.T) {
		r := New(io.NopCloser(strings.NewReader(msg1)), filterErr)
		err := iotest.TestReader(r, []byte(msg1))
		if err == nil {
			t.Error("Expected error, but returned without one")
		}
	})

	t.Run("", func(t *testing.T) {
		r := New(io.NopCloser(strings.NewReader(msg1)), filterPanic)
		err := iotest.TestReader(r, []byte(msg1))
		if err == nil {
			t.Error("Expected error, but returned without one")
		}
	})
}
