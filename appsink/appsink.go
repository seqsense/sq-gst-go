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

type SinkBufferHandler func([]byte, int)

type AppSink struct {
	element *gst.GstElement
	id      int32
}
type AppSinkHandlerInfo struct {
	handler SinkBufferHandler
}

var (
	handlers     = make(map[int32]*AppSinkHandlerInfo)
	handlerMutex sync.RWMutex
	idCnt        = int32(0)
)

func New(e *gst.GstElement, cb SinkBufferHandler) *AppSink {
	id := atomic.AddInt32(&idCnt, 1)
	s := &AppSink{
		element: e,
		id:      id,
	}
	handlerMutex.Lock()
	handlers[id] = &AppSinkHandlerInfo{
		handler: cb,
	}
	handlerMutex.Unlock()
	C.registerBufferHandler(e.UnsafePointer(), C.int(id))
	return s
}

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
