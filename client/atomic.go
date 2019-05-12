package client

import (
	"io"
	"sync/atomic"
	"unsafe"
)

type W struct {

}

func (w *W)Write(p []byte) (n int, err error)  {
	return 0, nil
}

func NewW() io.Writer {
	return &W{}
}

func LoadPiont()  {
	w := NewW()
	ww := (*W)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w))))
	ww.Write(nil)
}