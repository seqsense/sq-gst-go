// Copyright 2021 SEQSENSE, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
