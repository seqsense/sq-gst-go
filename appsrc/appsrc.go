package appsrc

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "appsrc.h"
import "C"

import (
	"unsafe"

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
	C.pushBuffer(s.element.UnsafePointer(), unsafe.Pointer(&buf[0]), C.int(len(buf)))
}

func (s *AppSrc) EOS() {
	C.sendEOS(s.element.UnsafePointer())
}
