package gstlaunch

import (
	"fmt"
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0 gstreamer-1.0 gstreamer-base-1.0
// #include "gstlaunch.h"
import "C"

func init() {
	C.init()
}

type GstLaunch struct {
	ctx     *C.Context
	quit    chan bool
	active  bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch)
}

var (
	cPointerMap          = make(map[int]*GstLaunch)
	cPointerMapIndex int = 0
)

func New(launch string) *GstLaunch {
	c_launch := C.CString(launch)
	defer C.free(unsafe.Pointer(c_launch))

	l := &GstLaunch{quit: make(chan bool, 1), active: false, cbEOS: nil, cbError: nil}
	cPointerMap[cPointerMapIndex] = l

	ctx := C.create(c_launch, C.int(cPointerMapIndex))
	if ctx == nil {
		panic("Failed to parse gst-launch text")
	}
	l.ctx = ctx

	cPointerMapIndex++

	runtime.SetFinalizer(l, finalizeGstLaunch)

	return l
}

func finalizeGstLaunch(s *GstLaunch) {
	C.free(unsafe.Pointer(s.ctx))
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
	s.active = true
	C.mainloopRun(s.ctx)
	s.quit <- true
	s.active = false
}

func (s *GstLaunch) Wait() {
	<-s.quit
}
func (s *GstLaunch) Kill() {
	C.mainloopKill(s.ctx)
}

func (s *GstLaunch) Active() bool {
	if s == nil {
		return false
	}
	return s.active
}
