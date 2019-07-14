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

import (
	"fmt"
	"runtime"
	"unsafe"
)

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

func (s GstState) String() string {
	switch s {
	case GST_STATE_VOID_PENDING:
		return "GST_STATE_VOID_PENDING"
	case GST_STATE_NULL:
		return "GST_STATE_NULL"
	case GST_STATE_READY:
		return "GST_STATE_READY"
	case GST_STATE_PAUSED:
		return "GST_STATE_PAUSED"
	case GST_STATE_PLAYING:
		return "GST_STATE_PLAYING"
	default:
		return fmt.Sprintf("Unknonw GstState (%d)", int(s))
	}
}

func NewGstElement(p unsafe.Pointer) *GstElement {
	e := &GstElement{p: p}
	runtime.SetFinalizer(e, finalizeGstElement)
	return e
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
