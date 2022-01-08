// iofilter can help filter data arriving from an io.ReadCloser
// It allows wraping an existing io.ReadCloser provider and filter
// the data before exposing it it up the chain.
package iofilter

import (
	"fmt"
	"io"
	"time"
)

// An Iofilter object maintining internal buffers and state
type Iofilter struct {
	inBuf      []byte
	outBuf     []byte
	bufChan    chan []byte
	bufs       [][]byte
	inBubIndex uint
	numBufs    uint
	sizeBuf    uint
	src        io.ReadCloser
	filter     func(buf []byte) error
}

// Create a New iofilter to wrap an existing provider of an io.ReadCloser interface
// The new iofilter will expose an io.ReadCloser interface
// The data will be sent to filter before it is delivered
// The optional params may include a two integer parameter indicating:
// 1. The number of buffers which may be at least 3 (default is 3)
// 2. The size of the buffers (default is 8192)
// A goroutine will be initiatd to wait on the original provider Read interface
// and deliver the data to the Readwer using an internal channel
func New(src io.ReadCloser, filter func(buf []byte) error, params ...uint) (iof *Iofilter) {
	var numBufs, sizeBuf uint
	fmt.Printf("params: %v\n", params)
	switch len(params) {
	case 0:
		numBufs = 3
		sizeBuf = 8192
	case 1:
		numBufs = params[0]
		if numBufs < 3 {
			numBufs = 3
		}
		sizeBuf = 8192
	case 2:
		numBufs = params[0]
		if numBufs < 3 {
			numBufs = 3
		}
		sizeBuf = params[1]
		if sizeBuf < 1 {
			sizeBuf = 1
		}
	default:
		panic("too many params in newStream")
	}

	iof = new(Iofilter)
	iof.numBufs = numBufs
	iof.sizeBuf = sizeBuf
	iof.filter = filter
	iof.src = src

	// create s.numBufs buffers
	iof.bufs = make([][]byte, iof.numBufs)
	for i := uint(0); i < iof.numBufs; i++ {
		iof.bufs[i] = make([]byte, iof.sizeBuf)
	}

	// we will maintain a maximum of s.numBufs-2 in s.bufChan + one buffer in s.inBuf + one buffer s.outBuf
	iof.bufChan = make(chan []byte, iof.numBufs-2)
	iof.inBuf = iof.bufs[0]

	// start serving the io
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				fmt.Printf("(iof *iofilter) Gorutine recovering from panic... %v\n", recovered)
			}

			// We close the internal channel to signal to Read() that we are done
			close(iof.bufChan)

			// Should we also close the source?
			// Did the source not reported an error?
			// Are we expected now to close it?
			// Seems just wrong.
			// iof.closeSrc()
		}()

		var n int
		var err error
		for {
			//fmt.Printf("(iof *iofilter) Gorutine Reading...\n")
			n, err = iof.src.Read(iof.inBuf)
			if n > 0 {
				//fmt.Printf("(iof *iofilter) Gorutine read %d bytes\n", n)
				err = iof.filter(iof.inBuf[:n])
				if err != nil {
					//fmt.Printf("(iof *iofilter) Gorutine filter blocked: %v\n", err)
					return
				}
				iof.bufChan <- iof.inBuf[:n]
				// ok, we now have a maximum of s.numBufs-2 in s.bufChan + one buffer s.outBuf
				// this means we have one free buffer to give to s.inBuf
				iof.inBubIndex = (iof.inBubIndex + 1) % iof.numBufs
				iof.inBuf = iof.bufs[iof.inBubIndex]
			} else {
				if err == nil {
					fmt.Printf("(iof *iofilter) Gorutine read no bytes, err is nil!\n")
					// hey, this io.Read interface is not doing as recommended!
					// "Implementations of Read are discouraged from returning a zero byte count with a nil error"
					// "Callers should treat a return of 0 and nil as indicating that nothing happened"
					// But even if nothing happened, we should not just abuse the CPU with an endless loop..
					time.Sleep(100 * time.Millisecond)
				}
			}
			if err != nil {
				if err.Error() != "EOF" {
					fmt.Printf("(iof *iofilter) Gorutine err %v\n", err)
				}
				return
			}
		}
	}()

	return
}

// The io.Read interface of the iofilter
func (iof *Iofilter) Read(dest []byte) (n int, err error) {
	//fmt.Printf("(iof *iofilter) Read\n")
	var opened bool
	err = nil
	// Do we have bytes in our current buffer?
	if len(iof.outBuf) == 0 {
		// Block until data arrives
		if iof.outBuf, opened = <-iof.bufChan; !opened {
			err = io.EOF
			n = 0
			//fmt.Printf("(iof *iofilter) Read Ended with io.EOF\n")
			return
		}
	}
	n = copy(dest, iof.outBuf)
	// We copied n bytes, lets skip them for next time
	iof.outBuf = iof.outBuf[n:]
	//fmt.Printf("(iof *iofilter) Read Ended after reading %d bytes\n", n)
	return
}

// The io.Close interface of the iofilter
func (iof *Iofilter) Close() error {
	//fmt.Printf("(iof *iofilter) Close\n")
	iof.closeSrc()
	return nil
}

func (iof *Iofilter) closeSrc() error {
	// There seem to be no standart convension about closing
	// Some may require it..
	// Others may alwaysb allow it..
	// Yet there are those who whould panic if closing when already closed..
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(iof *iofilter) recovering from panic during iof.src.Close() %v\n", recovered)
		}
	}()
	iof.src.Close()
	return nil
}
