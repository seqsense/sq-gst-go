package appsrc

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "appsrc.h"
import "C"

import (
	"unsafe"

	gst "github.com/seqsense/sq-gst-go"
)

// AppSrc is a wrapper of GStreamer AppSrc element.
type AppSrc struct {
	element *gst.Element
}

// New creates a GStreamer AppSrc element wrapper.
func New(e *gst.Element) *AppSrc {
	s := &AppSrc{
		element: e,
	}
	return s
}

// PushBuffer sends a buffer to the AppSrc.
func (s *AppSrc) PushBuffer(buf []byte) {
	C.pushBuffer(s.element.UnsafePointer(), unsafe.Pointer(&buf[0]), C.int(len(buf)))
}

// EOS sends end-of-stream message to the AppSrc.
func (s *AppSrc) EOS() {
	C.sendEOS(s.element.UnsafePointer())
}
