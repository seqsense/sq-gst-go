package types

// #cgo pkg-config: gstreamer-1.0
// #include <gst/gst.h>
// void unrefElement(void* element)
// {
//   gst_object_unref(element);
// }
import "C"

import "unsafe"

type GstElement struct {
	p unsafe.Pointer
}

func NewGstElement(p unsafe.Pointer) *GstElement {
	return &GstElement{p: p}
}

func finalizeGstElement(s *GstElement) {
	C.unrefElement(s.UnsafePointer())
}

func (s *GstElement) UnsafePointer() unsafe.Pointer {
	return s.p
}
