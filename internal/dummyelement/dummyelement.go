package dummyelement

import (
	"unsafe"
)

// #cgo pkg-config: gobject-2.0 gstreamer-1.0 gstreamer-base-1.0
// #include "gst/gst.h"
// void init()
// {
//   int argc = 1;
//   char* exec_name = "rtsp_receiver";
//   char** argv = &exec_name;
//   gst_init(&argc, &argv);
// }
// GstElement* newElement()
// {
//   return gst_element_factory_make("fakesink", "fakesink");
// }
import "C"

func init() {
	C.init()
}

// New returns dummy GstElement pointer. This is for internal testing.
func New() unsafe.Pointer {
	return unsafe.Pointer(C.newElement())
}
