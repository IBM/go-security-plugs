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
	inBufIndex uint
	numBufs    uint
	sizeBuf    uint
	src        io.ReadCloser
	filter     func(buf []byte, state interface{})
	state      interface{}
	done       chan bool
}

// Create a New iofilter to wrap an existing provider of an io.ReadCloser interface
// The new iofilter will expose an io.ReadCloser interface
// The data will be sent to filter before it is delivered
// The optional params may include a two integer parameter indicating:
// 1. The number of buffers which may be at least 3 (default is 3)
// 2. The size of the buffers (default is 8192)
// A goroutine will be initiatd to wait on the original provider Read interface
// and deliver the data to the Readwer using an internal channel
func New(src io.ReadCloser, filter func(buf []byte, state interface{}), state interface{}, params ...uint) (iof *Iofilter) {
	var numBufs, sizeBuf uint
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
	iof.state = state
	iof.done = make(chan bool)
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
		var n int
		var err error
		for err == nil {
			//fmt.Printf("(iof *iofilter) Gorutine Reading...\n")
			n, err = iof.readFromSrc()
			if n > 0 { // we have data
				//fmt.Printf("(iof *iofilter) Gorutine read %d bytes\n", n)
				iof.filterData(iof.inBuf[:n])

				iof.bufChan <- iof.inBuf[:n]
				// ok, we now have a maximum of s.numBufs-2 in s.bufChan + one buffer s.outBuf
				// this means we have one free buffer to give to s.inBuf
				iof.inBufIndex = (iof.inBufIndex + 1) % iof.numBufs
				iof.inBuf = iof.bufs[iof.inBufIndex]
			} else { // no data
				if err == nil { // no data and no err.... bad, bad writter!!
					fmt.Printf("(iof *iofilter) Gorutine read no bytes, err is nil!\n")
					// hey, this io.Read interface is not doing as recommended!
					// "Implementations of Read are discouraged from returning a zero byte count with a nil error"
					// "Callers should treat a return of 0 and nil as indicating that nothing happened"
					// But even if nothing happened, we should not just abuse the CPU with an endless loop..
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
		if err.Error() != "EOF" {
			fmt.Printf("(iof *iofilter) Gorutine err %v\n", err)
		} else {
			//fmt.Printf("(iof *iofilter) reached EOF in reader!\n")
		}

		iof.closeChannel()

		close(iof.done)
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
		if iof.outBuf, opened = <-iof.bufChan; !opened && iof.outBuf == nil {
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
	// We ignore close from any of the readers - we close when the source closes

	//fmt.Printf("(iof *iofilter) Close\n")
	//iof.closeSrc()
	return nil
}

/*
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
*/

func (iof *Iofilter) WaitTillDone() {
	<-iof.done
}

func (iof *Iofilter) readFromSrc() (n int, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(iof *iofilter) readFromSrc recovering from panic... %v\n", recovered)

			// We close the internal channel to signal from the src to readers that we are done
			close(iof.bufChan)

			n = 0
			err = io.EOF
		}
	}()
	n, err = iof.src.Read(iof.inBuf)
	return n, err
}

func (iof *Iofilter) filterData(buf []byte) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(iof *iofilter) filterData recovering from panic... %v\n", recovered)
		}
	}()
	iof.filter(buf, iof.state)
}

func (iof *Iofilter) closeChannel() {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(iof *Iofilter) closeChannel recovering from panic... %v\n", recovered)
		}
	}()
	//fmt.Printf("(iof *Iofilter) closeChannel ! \n")
	close(iof.bufChan)
}
