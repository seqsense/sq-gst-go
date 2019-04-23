package gst

// #cgo pkg-config: gstreamer-1.0
// #include <gst/gst.h>
// void unrefElement(void* element)
// {
//   gst_object_unref(element);
// }
// GstState getElementState(void* element)
// {
//   return GST_STATE(element);
// }
import "C"

import "unsafe"

type GstElement struct {
	p unsafe.Pointer
}

type GstState uint8

const (
	GST_STATE_VOID_PENDING GstState = iota
	GST_STATE_NULL
	GST_STATE_READY
	GST_STATE_PAUSED
	GST_STATE_PLAYING
)

func NewGstElement(p unsafe.Pointer) *GstElement {
	return &GstElement{p: p}
}

func finalizeGstElement(s *GstElement) {
	C.unrefElement(s.UnsafePointer())
}

func (s *GstElement) UnsafePointer() unsafe.Pointer {
	return s.p
}

func (s *GstElement) State() GstState {
	return GstState(C.getElementState(s.p))
}
