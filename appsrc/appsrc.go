package appsrc

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "appsrc.h"
import "C"

import (
	"runtime"

	"github.com/seqsense/sq-gst-go/types"
)

type AppSrc struct {
	element *types.GstElement
}

func New(e *types.GstElement) *AppSrc {
	s := &AppSrc{
		element: e,
	}
	runtime.SetFinalizer(s, finalizeAppSrc)
	return s
}

func finalizeAppSrc(s *AppSrc) {
	C.unrefElement(s.element.UnsafePointer())
}

func (s *AppSrc) PushBuffer(buf []byte) {
	c_buf := C.CBytes(buf)
	defer C.free(c_buf)
	C.pushBuffer(s.element.UnsafePointer(), c_buf, C.int(len(buf)))
}
