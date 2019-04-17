package appsrc

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "appsrc.h"
import "C"

import (
	gst "github.com/seqsense/sq-gst-go"
)

type AppSrc struct {
	element *gst.GstElement
}

func New(e *gst.GstElement) *AppSrc {
	s := &AppSrc{
		element: e,
	}
	return s
}

func (s *AppSrc) PushBuffer(buf []byte) {
	c_buf := C.CBytes(buf)
	defer C.free(c_buf)
	C.pushBuffer(s.element.UnsafePointer(), c_buf, C.int(len(buf)))
}
