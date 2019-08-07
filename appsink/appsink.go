package appsink

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "appsink.h"
import "C"

import (
	"log"
	"sync"
	"sync/atomic"
	"unsafe"

	gst "github.com/seqsense/sq-gst-go"
)

// BufferHandler is a stream buffer handler callback type.
type BufferHandler func([]byte, int)

// AppSink is a wrapper of GStreamer AppSink element.
type AppSink struct {
	element *gst.Element
	id      int32
}

type handlerInfo struct {
	handler BufferHandler
}

var (
	handlers     = make(map[int32]*handlerInfo)
	handlerMutex sync.RWMutex
	idCnt        = int32(0)
)

// New creates a GStreamer AppSink element wrapper.
func New(e *gst.Element, cb BufferHandler) *AppSink {
	id := atomic.AddInt32(&idCnt, 1)
	s := &AppSink{
		element: e,
		id:      id,
	}
	handlerMutex.Lock()
	handlers[id] = &handlerInfo{
		handler: cb,
	}
	handlerMutex.Unlock()
	C.registerBufferHandler(e.UnsafePointer(), C.int(id))
	return s
}

// Close stops AppSink handling and free resource.s
func (s *AppSink) Close() {
	handlerMutex.Lock()
	delete(handlers, s.id)
	handlerMutex.Unlock()
}

//export goBufferHandler
func goBufferHandler(p unsafe.Pointer, len, samples, id C.int) {
	handlerMutex.RLock()
	h, ok := handlers[int32(id)]
	handlerMutex.RUnlock()
	if ok {
		h.handler(C.GoBytes(p, len), int(samples))
	} else {
		log.Printf("Unhandled buffer received (id: %d)", int(id))
	}
}
