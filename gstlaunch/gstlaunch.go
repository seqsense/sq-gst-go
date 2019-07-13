package gstlaunch

import (
	"context"
	"errors"
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
	cCtx    *C.Context
	active  bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch)
	index   int
	cbLock  sync.Mutex
	done    chan (struct{})
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
		done:    make(chan struct{}),
	}

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

func finalizeGstLaunch(l *GstLaunch) {
	cPointerMapMutex.Lock()
	defer cPointerMapMutex.Unlock()

	C.pipelineUnref(l.cCtx)
	C.free(unsafe.Pointer(l.cCtx))
	delete(cPointerMap, l.index)
}

func (l *GstLaunch) RegisterErrorCallback(f func(*GstLaunch)) {
	l.cbError = f
}

func (l *GstLaunch) RegisterEOSCallback(f func(*GstLaunch)) {
	l.cbEOS = f
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

//export goCbState
func goCbState(i C.int, oldState, newState, pendingState C.uint) {
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	switch gst.GstState(newState) {
	case gst.GST_STATE_PLAYING:
		l.active = true
	default:
		l.active = false
	}
}

func (l *GstLaunch) Run(ctx context.Context) error {
	l.Start()
	return l.Wait(ctx)
}
func (l *GstLaunch) Start() {
	l.active = true
	C.pipelineStart(l.cCtx)
}
func (l *GstLaunch) Kill() {
	C.pipelineStop(l.cCtx)
	close(l.done)
}
func (l *GstLaunch) Wait(ctx context.Context) error {
	select {
	case <-l.done:
		return nil
	case <-ctx.Done():
		return errors.New("wait timeout")
	}
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
