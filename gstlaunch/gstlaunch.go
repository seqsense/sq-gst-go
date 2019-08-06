package gstlaunch

import (
	"fmt"
	"log"
	"os"
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
	closed  bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch)
	cbState func(*GstLaunch, gst.GstState, gst.GstState, gst.GstState)
	index   int
	cbLock  sync.Mutex
}

var (
	cPointerMap          = make(map[int]*GstLaunch)
	cPointerMapIndex int = 0
	cPointerMapMutex     = sync.RWMutex{}
	closedErr            = fmt.Errorf("pipeline is closed")
)

func New(launch string) *GstLaunch {
	c_launch := C.CString(launch)
	defer C.free(unsafe.Pointer(c_launch))

	l := &GstLaunch{
		active:  false,
		cbEOS:   nil,
		cbError: nil,
		cbState: nil,
		cbLock:  sync.Mutex{},
	}

	cPointerMapMutex.Lock()
	id := cPointerMapIndex
	cPointerMap[id] = l
	cPointerMapIndex++
	cPointerMapMutex.Unlock()

	l.index = id

	cCtx := C.create(c_launch, C.int(id))
	if cCtx == nil {
		panic("Failed to parse gst-launch text")
	}
	l.cCtx = cCtx
	return l
}

func (l *GstLaunch) unref() error {
	if l.closed {
		return closedErr
	}
	l.closed = true
	C.pipelineUnref(l.cCtx)

	cPointerMapMutex.Lock()
	delete(cPointerMap, l.index)
	cPointerMapMutex.Unlock()
	return nil
}

func (l *GstLaunch) RegisterErrorCallback(f func(*GstLaunch)) error {
	if l.closed {
		return closedErr
	}
	l.cbError = f
	return nil
}

func (l *GstLaunch) RegisterEOSCallback(f func(*GstLaunch)) error {
	if l.closed {
		return closedErr
	}
	l.cbEOS = f
	return nil
}

func (l *GstLaunch) RegisterStateCallback(f func(*GstLaunch, gst.GstState, gst.GstState, gst.GstState)) error {
	if l.closed {
		return closedErr
	}
	l.cbState = f
	return nil
}

//export goCbEOS
func goCbEOS(i C.int) {
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		log.Printf("Failed to map pointer from cgo func (EOS message, %d)", int(i))
		return
	}
	if l.cbEOS != nil {
		l.cbLock.Lock()
		l.cbEOS(l)
		l.cbLock.Unlock()
	}
}

//export goCbError
func goCbError(i C.int) {
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		log.Printf("Failed to map pointer from cgo func (error message, %d)", int(i))
		return
	}
	if l.cbError != nil {
		l.cbLock.Lock()
		l.cbError(l)
		l.cbLock.Unlock()
	}
}

//export goCbState
func goCbState(i C.int, oldState, newState, pendingState C.uint) {
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		log.Printf("Failed to map pointer from cgo func (state message, %d)", int(i))
		return
	}
	l.setState(gst.GstState(oldState), gst.GstState(newState), gst.GstState(pendingState))
}

func (l *GstLaunch) setState(o, n, p gst.GstState) {
	if l.cbState != nil {
		l.cbLock.Lock()
		l.cbState(l, o, n, p)
		l.cbLock.Unlock()
	}
	switch n {
	case gst.GST_STATE_PLAYING:
		l.active = true
	case gst.GST_STATE_NULL:
		l.unref()
		l.active = false
	default:
		l.active = false
	}
}

func (l *GstLaunch) Start() error {
	if l.closed {
		return closedErr
	}
	C.pipelineStart(l.cCtx)
	return nil
}
func (l *GstLaunch) Kill() error {
	if l.closed {
		return closedErr
	}
	C.pipelineStop(l.cCtx)
	// Transition to GST_STATE_NULL is guaranteed to be synchronous and message is no longer reachable.
	l.setState(gst.GST_STATE_READY, gst.GST_STATE_NULL, gst.GST_STATE_VOID_PENDING)
	return nil
}
func (l *GstLaunch) Active() bool {
	if l == nil {
		return false
	}
	if l.closed {
		return false
	}
	return l.active
}

func (l *GstLaunch) GetElement(name string) (*gst.GstElement, error) {
	if l.closed {
		return nil, closedErr
	}
	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))

	e := C.getElement(l.cCtx, c_name)
	if e == nil {
		return nil, fmt.Errorf("Failed to get %s", name)
	}
	return gst.NewGstElement(unsafe.Pointer(e)), nil
}
