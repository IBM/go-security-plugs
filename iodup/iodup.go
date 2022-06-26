// iodup can help dulicate data arriving from an io.ReadCloser
// It allows wraping an existing io.ReadCloser provider and duplciate it to n readers
package iodup

import (
	"fmt"
	"io"
	"time"
)

// An Iodup object maintining internal buffers and state
type Out struct {
	outBuf  []byte
	bufChan chan []byte
}

type Iodup struct {
	inBuf      []byte
	Output     []Out
	bufs       [][]byte
	inBufIndex uint
	numBufs    uint
	numOutputs uint
	sizeBuf    uint
	src        io.ReadCloser
}

// Create a New iodup to wrap an existing provider of an io.ReadCloser interface
// The new iodup.out[] will expose an io.ReadCloser interface
// The optional params may include a two integer parameter indicating:
// 1. The number of outputs (defualt is 2)
// 2. The number of buffers which may be at least 3 (default is 1024)
// 3. The size of the buffers (default is 8192)
// A goroutine will be initiatd to wait on the original provider Read interface
// and deliver the data to the Readwer using an internal channel
func New(src io.ReadCloser, params ...uint) (iod *Iodup) {
	var numOutputs, numBufs, sizeBuf uint
	fmt.Printf("params: %v\n", params)
	switch len(params) {
	case 0:
		numOutputs = 2
		numBufs = 1024
		sizeBuf = 8192
	case 1:
		numOutputs = params[0]
		if numOutputs < 2 {
			numOutputs = 2
		}
		numBufs = 1024
		sizeBuf = 8192
	case 2:
		numOutputs = params[0]
		if numOutputs < 2 {
			numOutputs = 2
		}
		numBufs = params[1]
		if numBufs < 3 {
			numBufs = 1024
		}
		sizeBuf = 8192
	case 3:
		numOutputs = params[0]
		if numOutputs < 2 {
			numOutputs = 2
		}
		numBufs = params[1]
		if numBufs < 3 {
			numBufs = 1024
		}
		sizeBuf = params[2]
		if sizeBuf < 1 {
			sizeBuf = 1
		}
	default:
		panic("too many params in newStream")
	}

	iod = new(Iodup)
	iod.numOutputs = numOutputs
	iod.numBufs = numBufs
	iod.sizeBuf = sizeBuf
	iod.src = src

	// create s.numOutputs outputs
	iod.Output = make([]Out, iod.numOutputs)
	for j := uint(0); j < iod.numOutputs; j++ {
		// we will maintain a maximum of s.numBufs-2 in s.bufChan + one buffer in s.inBuf + one buffer s.outBuf
		iod.Output[j].bufChan = make(chan []byte, iod.numBufs-2)
	}
	iod.bufs = make([][]byte, iod.numBufs)
	for i := uint(0); i < iod.numBufs; i++ {
		iod.bufs[i] = make([]byte, iod.sizeBuf)
	}
	iod.inBuf = iod.bufs[0]

	// start serving the io
	go func() {
		var n int
		var err error
		for err == nil {
			n, err = iod.readFromSrc()
			if n > 0 { // we have data
				iod.forwardToOut(iod.inBuf[:n])

				// ok, we now have a maximum of s.numBufs-2 in s.bufChan + one buffer s.outBuf
				// this means we have one free buffer to give to s.inBuf
				iod.inBufIndex = (iod.inBufIndex + 1) % iod.numBufs
				iod.inBuf = iod.bufs[iod.inBufIndex]
			} else { // no data
				if err == nil { // no data and no err.... bad, bad writter!!
					//fmt.Printf("(iof *iofilter) Gorutine read no bytes, err is nil!\n")
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

		for j := uint(0); j < iod.numOutputs; j++ {
			iod.Output[j].closeChannel()
		}
	}()

	return
}

func (iod *Iodup) forwardToOut(buf []byte) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(iof *iofilter) forwardToOut recovering from panic... %v\n", recovered)
		}

		// we never close bufChan from the receiver side, so we should never panic here!
		// closing the source is not a great idea...

		// We close the internal channel to signal to Read() that we are done
		//for j := uint(0); j < iof.numOutputs; j++ {
		//	close(iof.out[j].bufChan)
		//}

		//iod.closeSrc()
	}()

	//fmt.Printf("(iod *Iodup) Gorutine forward %d bytes\n", len(buf))
	for j := uint(0); j < iod.numOutputs; j++ {
		iod.Output[j].bufChan <- buf
	}
}
func (iod *Iodup) readFromSrc() (n int, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(iof *iofilter) readFromSrc recovering from panic... %v\n", recovered)

			// We close the internal channel to signal from the src to readers that we are done
			for j := uint(0); j < iod.numOutputs; j++ {
				close(iod.Output[j].bufChan)
			}
			n = 0
			err = io.EOF
		}
	}()
	//fmt.Printf("(iof *Iodup) Gorutine readFromSrc Reading...\n")
	n, err = iod.src.Read(iod.inBuf)
	//fmt.Printf("(iof *Iodup) Gorutine readFromSrc returning %d err %v\n", n, err)

	return n, err
}

// The io.Read interface of the iofilter
func (out *Out) Read(dest []byte) (n int, err error) {
	//fmt.Printf("(out *Out) Read while len(out.outBuf) is %d\n", len(out.outBuf))
	var opened bool
	err = nil
	// Do we have bytes in our current buffer?
	if len(out.outBuf) == 0 {
		// Block until data arrives
		if out.outBuf, opened = <-out.bufChan; !opened && out.outBuf == nil {
			//fmt.Printf("(out *Out) Read out.outBuf %v opened %v \n", out.outBuf, opened)

			err = io.EOF
			n = 0
			//fmt.Printf("(out *Out) Read Ended  - channel is closed! ending with io.EOF\n")
			return
		}
		//fmt.Printf("(out *Out) Read with new buffer len(out.outBuf) is %d\n", len(out.outBuf))
	}

	n = copy(dest, out.outBuf)
	// We copied n bytes, lets skip them for next time
	out.outBuf = out.outBuf[n:]
	//fmt.Printf("(out *Out) Read Ended after reading %d bytes\n", n)
	return
}

// The io.Close interface of the iofilter
func (out *Out) Close() error {
	// We ignore close from any of the readers - we close when the source closes

	//fmt.Printf("(iof *iofilter) Close\n")
	//close(out.bufChan)
	//iof.closeSrc()
	return nil
}
func (out *Out) closeChannel() {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("(out *Out) closeChannel recovering from panic... %v\n", recovered)
		}
	}()
	//fmt.Printf("(out *Out) closeChannel ! \n")
	close(out.bufChan)
}

/*   closing teh src is not a great idea
func (iof *Iodup) closeSrc() error {
	// There seem to be no standart convension about closing
	// Some may require it..
	// Others may always allow it..
	// Yet there are those who whould panic if closing when already closed..
	//defer func() {
	//	if recovered := recover(); recovered != nil {
	//		fmt.Printf("(iof *iofilter) recovering from panic during iof.src.Close() %v\n", recovered)
	//	}
	//}()
	//iof.src.Close()
	return nil
}
*/
