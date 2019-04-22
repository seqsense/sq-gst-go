package appsink

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "appsink.h"
import "C"

import (
	"log"
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
	handlers map[int32]*AppSinkHandlerInfo
	idCnt    int32
)

func init() {
	idCnt = 0
	handlers = make(map[int32]*AppSinkHandlerInfo)
}

func New(e *gst.GstElement, cb SinkBufferHandler) *AppSink {
	id := atomic.AddInt32(&idCnt, 1)
	s := &AppSink{
		element: e,
		id:      id,
	}
	handlers[id] = &AppSinkHandlerInfo{
		handler: cb,
	}
	C.registerBufferHandler(e.UnsafePointer(), C.int(id))
	return s
}

func (s *AppSink) Close() {
	delete(handlers, s.id)
}

//export goBufferHandler
func goBufferHandler(p unsafe.Pointer, len, samples, id C.int) {
	if h, ok := handlers[int32(id)]; ok {
		h.handler(C.GoBytes(p, len), int(samples))
	} else {
		log.Printf("Unhandled buffer received (id: %d)", int(id))
	}
}