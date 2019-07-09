package gstlaunch

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"unsafe"

	gst "github.com/seqsense/sq-gst-go"
)

// #cgo pkg-config: gobject-2.0 gstreamer-1.0 gstreamer-base-1.0
// #include "gstlaunch.h"
import "C"

func init() {
	n := C.CString(os.Args[0])
	defer C.free(unsafe.Pointer(n))
	C.init(n)

	go C.runMainloop()
}

type GstLaunch struct {
	ctx     context.Context
	cancel  func()
	cCtx    *C.Context
	active  bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch)
	index   int
}

var (
	cPointerMap          = make(map[int]*GstLaunch)
	cPointerMapIndex int = 0
)

func New(launch string) *GstLaunch {
	c_launch := C.CString(launch)
	defer C.free(unsafe.Pointer(c_launch))

	l := &GstLaunch{active: false, cbEOS: nil, cbError: nil, index: cPointerMapIndex}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	cPointerMap[cPointerMapIndex] = l

	cCtx := C.create(c_launch, C.int(cPointerMapIndex))
	if cCtx == nil {
		panic("Failed to parse gst-launch text")
	}
	l.cCtx = cCtx

	cPointerMapIndex++

	runtime.SetFinalizer(l, finalizeGstLaunch)

	return l
}

func finalizeGstLaunch(s *GstLaunch) {
	C.free(unsafe.Pointer(s.cCtx))
	delete(cPointerMap, s.index)
}

func (s *GstLaunch) RegisterErrorCallback(f func(*GstLaunch)) {
	s.cbError = f
}

func (s *GstLaunch) RegisterEOSCallback(f func(*GstLaunch)) {
	s.cbEOS = f
}

//export goCbEOS
func goCbEOS(i C.int) {
	s, ok := cPointerMap[int(i)]
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	if s.cbEOS != nil {
		s.cbEOS(s)
	}
}

//export goCbError
func goCbError(i C.int) {
	s, ok := cPointerMap[int(i)]
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	if s.cbError != nil {
		s.cbError(s)
	}
}

func (s *GstLaunch) Run() {
	s.Start()
	<-s.ctx.Done()
}
func (s *GstLaunch) Start() {
	s.active = true
	C.pipelineStart(s.cCtx)
}
func (s *GstLaunch) Wait() {
	<-s.ctx.Done()
}
func (s *GstLaunch) Kill() {
	s.active = false
	C.pipelineKill(s.cCtx)
	s.cancel()
}

func (s *GstLaunch) Active() bool {
	if s == nil {
		return false
	}
	return s.active
}

func (s *GstLaunch) GetElement(name string) (*gst.GstElement, error) {
	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))

	e := C.getElement(s.cCtx, c_name)
	if e == nil {
		return nil, fmt.Errorf("Failed to get %s", name)
	}
	return gst.NewGstElement(unsafe.Pointer(e)), nil
}
