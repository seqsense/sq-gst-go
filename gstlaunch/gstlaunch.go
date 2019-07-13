package gstlaunch

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
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
	cbLock  sync.Mutex
}

var (
	cPointerMap          = make(map[int]*GstLaunch)
	cPointerMapIndex int = 0
	cPointerMapMutex     = sync.RWMutex{}
)

func New(launch string) *GstLaunch {
	c_launch := C.CString(launch)
	defer C.free(unsafe.Pointer(c_launch))

	l := &GstLaunch{
		active:  false,
		cbEOS:   nil,
		cbError: nil,
		index:   cPointerMapIndex,
		cbLock:  sync.Mutex{},
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())

	cPointerMapMutex.Lock()
	cPointerMap[cPointerMapIndex] = l
	cPointerMapMutex.Unlock()

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
	cPointerMapMutex.Lock()
	defer cPointerMapMutex.Unlock()

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
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	l.cbLock.Lock()
	defer l.cbLock.Unlock()
	if l.cbEOS != nil {
		l.cbEOS(l)
	}
}

//export goCbError
func goCbError(i C.int) {
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	l.cbLock.Lock()
	defer l.cbLock.Unlock()
	if l.cbError != nil {
		l.cbError(l)
	}
}

func (l *GstLaunch) Run() {
	l.Start()
	<-l.ctx.Done()
}
func (l *GstLaunch) Start() {
	l.active = true
	C.pipelineStart(l.cCtx)
}
func (l *GstLaunch) Wait() {
	<-l.ctx.Done()
}
func (l *GstLaunch) Kill() {
	l.active = false
	C.pipelineKill(l.cCtx)
	l.cancel()
}

func (l *GstLaunch) Active() bool {
	if l == nil {
		return false
	}
	return l.active
}

func (l *GstLaunch) GetElement(name string) (*gst.GstElement, error) {
	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))

	e := C.getElement(l.cCtx, c_name)
	if e == nil {
		return nil, fmt.Errorf("Failed to get %s", name)
	}
	return gst.NewGstElement(unsafe.Pointer(e)), nil
}
