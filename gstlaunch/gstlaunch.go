// Copyright 2021 SEQSENSE, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gstlaunch

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
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
}

// GstLaunch is a wrapper of GstPipeline structured from launch string.
type GstLaunch struct {
	cCtx    *C.Context
	active  atomic.Value // bool
	closed  atomic.Value // bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch, *gst.Element, string, string)
	cbState func(*GstLaunch, gst.State, gst.State, gst.State)
	index   int
	mu      sync.RWMutex
}

var (
	cPointerMapIndex int
	cPointerMap      = make(map[int]*GstLaunch)
	cPointerMapMutex = sync.RWMutex{}
	numCtx           int
	numCtxMutex      = sync.RWMutex{}
	errClosed        = fmt.Errorf("pipeline is closed")
)

// New creates a new GstPipeline wrapper from launch string.
func New(launch string) (*GstLaunch, error) {
	cLaunch := C.CString(launch)
	defer C.free(unsafe.Pointer(cLaunch))

	l := &GstLaunch{
		cbEOS:   nil,
		cbError: nil,
		cbState: nil,
		mu:      sync.RWMutex{},
	}
	l.active.Store(false)
	l.closed.Store(false)

	cPointerMapMutex.Lock()
	id := cPointerMapIndex
	cPointerMap[id] = l
	cPointerMapIndex++
	cPointerMapMutex.Unlock()

	l.index = id

	cCtx := C.create(cLaunch, C.int(id))
	if cCtx == nil {
		return nil, fmt.Errorf("Failed to create gstlaunch pipeline")
	}
	l.cCtx = cCtx

	numCtxMutex.Lock()
	numCtx++
	numCtxMutex.Unlock()
	return l, nil
}

func getNumCtx() int {
	numCtxMutex.RLock()
	defer numCtxMutex.RUnlock()
	return numCtx
}

func (l *GstLaunch) unref() error {
	if l.closed.Load().(bool) {
		return errClosed
	}
	l.closed.Store(true)
	go func() {
		time.Sleep(10 * time.Millisecond)
		C.pipelineUnref(l.cCtx)

		cPointerMapMutex.Lock()
		delete(cPointerMap, l.index)
		cPointerMapMutex.Unlock()

		// FIXME(at-wat): find more proper way to ensure no more handlers are called
		time.Sleep(time.Second)
		C.pipelineFree(l.cCtx)
		numCtxMutex.Lock()
		numCtx--
		numCtxMutex.Unlock()
	}()
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
func (l *GstLaunch) RegisterErrorCallback(f func(*GstLaunch, *gst.Element, string, string)) error {
	if l.closed.Load().(bool) {
		return errClosed
	}
	l.mu.Lock()
	l.cbError = f
	l.mu.Unlock()
	return nil
}

// RegisterEOSCallback registers EOS message handler callback.
func (l *GstLaunch) RegisterEOSCallback(f func(*GstLaunch)) error {
	if l.closed.Load().(bool) {
		return errClosed
	}
	l.mu.Lock()
	l.cbEOS = f
	l.mu.Unlock()
	return nil
}

// RegisterStateCallback registers state update message handler callback.
func (l *GstLaunch) RegisterStateCallback(f func(*GstLaunch, gst.State, gst.State, gst.State)) error {
	if l.closed.Load().(bool) {
		return errClosed
	}
	l.mu.Lock()
	l.cbState = f
	l.mu.Unlock()
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
	l.mu.RLock()
	cb := l.cbEOS
	l.mu.RUnlock()
	if cb != nil {
		cb(l)
	}
}

//export goCbError
func goCbError(i C.int, e unsafe.Pointer, msg *C.char, msgSize C.int, dbgInfo *C.char, dbgInfoSize C.int) {
	cPointerMapMutex.RLock()
	l, ok := cPointerMap[int(i)]
	cPointerMapMutex.RUnlock()
	if !ok {
		log.Printf("Failed to map pointer from cgo func (error message, %d)", int(i))
		return
	}
	l.mu.RLock()
	cb := l.cbError
	l.mu.RUnlock()

	msgGo := C.GoStringN(msg, msgSize)
	dbgInfoGo := ""
	if dbgInfo != nil {
		dbgInfoGo = C.GoStringN(dbgInfo, dbgInfoSize)
	}
	if cb != nil {
		C.refElement(e)
		cb(l, gst.NewElement(e), msgGo, dbgInfoGo)
	} else {
		log.Printf("Unhandled error message \"%s\":\n%s", msgGo, dbgInfoGo)
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
	l.mu.RLock()
	cb := l.cbState
	l.mu.RUnlock()
	if cb != nil {
		cb(l, o, n, p)
	}
	switch n {
	case gst.StatePlaying:
		l.active.Store(true)
	case gst.StateNull:
		l.unref()
		l.active.Store(false)
	default:
		l.active.Store(false)
	}
}

// Start makes the pipeline playing.
func (l *GstLaunch) Start() error {
	if l.closed.Load().(bool) {
		return errClosed
	}
	C.pipelineStart(l.cCtx)
	return nil
}

// Kill stops the pipeline and free resources.
func (l *GstLaunch) Kill() error {
	if l.closed.Load().(bool) {
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
	if l.closed.Load().(bool) {
		return false
	}
	return l.active.Load().(bool)
}

// GetElement finds GstElement by the name.
func (l *GstLaunch) GetElement(name string) (*gst.Element, error) {
	if l.closed.Load().(bool) {
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

// GetAllElements returns all GstElement in the pipeline.
func (l *GstLaunch) GetAllElements() ([]*gst.Element, error) {
	if l.closed.Load().(bool) {
		return nil, errClosed
	}
	var ret []*gst.Element
	es := C.getAllElements(l.cCtx)
	defer C.free(unsafe.Pointer(es))

	for i := 0; ; i++ {
		e := C.elementAt(es, C.int(i))
		if e == nil {
			break
		}
		ret = append(ret, gst.NewElement(unsafe.Pointer(e)))
	}
	return ret, nil
}
