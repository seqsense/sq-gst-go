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

// GstLaunch is a wrapper of GstPipeline structured from launch string.
type GstLaunch struct {
	cCtx    *C.Context
	active  bool
	closed  bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch)
	cbState func(*GstLaunch, gst.State, gst.State, gst.State)
	index   int
	cbLock  sync.Mutex
}

var (
	cPointerMapIndex int
	cPointerMap      = make(map[int]*GstLaunch)
	cPointerMapMutex = sync.RWMutex{}
	errClosed        = fmt.Errorf("pipeline is closed")
)

// New creates a new GstPipeline wrapper from launch string.
func New(launch string) (*GstLaunch, error) {
	cLaunch := C.CString(launch)
	defer C.free(unsafe.Pointer(cLaunch))

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

	cCtx := C.create(cLaunch, C.int(id))
	if cCtx == nil {
		return nil, fmt.Errorf("Failed to parse gst-launch text")
	}
	l.cCtx = cCtx
	return l, nil
}

func (l *GstLaunch) unref() error {
	if l.closed {
		return errClosed
	}
	l.closed = true
	C.pipelineUnref(l.cCtx)

	cPointerMapMutex.Lock()
	delete(cPointerMap, l.index)
	cPointerMapMutex.Unlock()
	return nil
}

// MustNew creates a new GstPipeline wrapper from launch string. It panics on fail.
func MustNew(launch string) *GstLaunch {
	l, err := New(launch)
	if err != nil {
		panic(err)
	}
	return l
}

// RegisterErrorCallback registers error message handler callback.
func (l *GstLaunch) RegisterErrorCallback(f func(*GstLaunch)) error {
	if l.closed {
		return errClosed
	}
	l.cbError = f
	return nil
}

// RegisterEOSCallback registers EOS message handler callback.
func (l *GstLaunch) RegisterEOSCallback(f func(*GstLaunch)) error {
	if l.closed {
		return errClosed
	}
	l.cbEOS = f
	return nil
}

// RegisterStateCallback registers state update message handler callback.
func (l *GstLaunch) RegisterStateCallback(f func(*GstLaunch, gst.State, gst.State, gst.State)) error {
	if l.closed {
		return errClosed
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
	l.setState(gst.State(oldState), gst.State(newState), gst.State(pendingState))
}

func (l *GstLaunch) setState(o, n, p gst.State) {
	if l.cbState != nil {
		l.cbLock.Lock()
		l.cbState(l, o, n, p)
		l.cbLock.Unlock()
	}
	switch n {
	case gst.StatePlaying:
		l.active = true
	case gst.StateNull:
		l.unref()
		l.active = false
	default:
		l.active = false
	}
}

// Start makes the pipeline playing.
func (l *GstLaunch) Start() error {
	if l.closed {
		return errClosed
	}
	C.pipelineStart(l.cCtx)
	return nil
}

// Kill stops the pipeline and free resources.
func (l *GstLaunch) Kill() error {
	if l.closed {
		return errClosed
	}
	C.pipelineStop(l.cCtx)
	// Transition to StateNULL is guaranteed to be synchronous and message is no longer reachable.
	l.setState(gst.StateReady, gst.StateNull, gst.StateVoidPending)
	return nil
}

// Active returns true if the pipeline is playing.
func (l *GstLaunch) Active() bool {
	if l == nil {
		return false
	}
	if l.closed {
		return false
	}
	return l.active
}

// GetElement finds GstElement by the name.
func (l *GstLaunch) GetElement(name string) (*gst.Element, error) {
	if l.closed {
		return nil, errClosed
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	e := C.getElement(l.cCtx, cName)
	if e == nil {
		return nil, fmt.Errorf("Failed to get %s", name)
	}
	return gst.NewElement(unsafe.Pointer(e)), nil
}
