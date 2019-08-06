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

// Element is a wrapper of GstElement.
type Element struct {
	p unsafe.Pointer
}

// State is a GStreamer element state.
type State uint8

const (
	// StateVoidPending states that the element don't have pending state.
	StateVoidPending State = iota
	// StateNull states that the element is initial or finalized state.
	StateNull
	// StateReady states that the element allocated resources.
	StateReady
	// StatePaused states that the element is paused and ready to accept data.
	StatePaused
	//StatePlaying states that the element is playing.
	StatePlaying
)

// String returns string representation of the State.
func (s State) String() string {
	switch s {
	case StateVoidPending:
		return "StateVoidPending"
	case StateNull:
		return "StateNull"
	case StateReady:
		return "StateReady"
	case StatePaused:
		return "StatePaused"
	case StatePlaying:
		return "StatePlaying"
	default:
		return fmt.Sprintf("Unknonw State (%d)", int(s))
	}
}

// NewElement creates a new GStreamer element wrapper from given raw pointer.
func NewElement(p unsafe.Pointer) *Element {
	e := &Element{p: p}
	runtime.SetFinalizer(e, finalizeElement)
	return e
}

func finalizeElement(s *Element) {
	C.unrefElement(s.UnsafePointer())
}

// UnsafePointer returns the raw pointer of the element.
func (s *Element) UnsafePointer() unsafe.Pointer {
	return s.p
}

// State returns the current state of the element.
func (s *Element) State() State {
	return State(C.getElementState(s.p))
}
