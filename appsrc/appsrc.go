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
