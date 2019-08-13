package gst

// #cgo pkg-config: gobject-2.0 gstreamer-1.0
// #include <stdlib.h>
// #include <glib-object.h>
// #include <gst/gst.h>
// void unrefElement(void* element)
// {
//   gst_object_unref(element);
// }
// GstState getElementState(void* element)
// {
//   return GST_STATE(element);
// }
// GValue* newGValue()
// {
//   GValue* value = malloc(sizeof(GValue));
//   GValue init = G_VALUE_INIT;
//   *value = init;
//   return value;
// }
// GValue* getProperty(void* element, const char* name)
// {
//   g_object_ref(element);
//   GObjectClass klass;
//   klass.g_type_class.g_type = G_OBJECT_TYPE(element);
//   GParamSpec* pspec = g_object_class_find_property(&klass, name);
//   if (pspec == NULL)
//   {
//     g_object_unref(element);
//     return NULL;
//   }
//   GValue* value = newGValue();
//   g_value_init(value, G_PARAM_SPEC_VALUE_TYPE(pspec));
//   g_object_get_property(element, name, value);
//   g_object_unref(element);
//   return value;
// }
// void setProperty(void* element, const char* name, GValue* value)
// {
//   g_object_set_property(element, name, value);
// }
// GType getValueType(GValue* value)
// {
//   return G_VALUE_TYPE(value);
// }
import "C"

import (
	"fmt"
	"reflect"
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

// GetProperty returns property of the element.
func (s *Element) GetProperty(name string) (interface{}, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	v := C.getProperty(s.UnsafePointer(), cName)
	if v == nil {
		return nil, fmt.Errorf("Property not found")
	}
	defer C.free(unsafe.Pointer(v))

	t := C.getValueType(v)
	switch t {
	case C.G_TYPE_INT:
		return int(C.g_value_get_int(v)), nil
	case C.G_TYPE_UINT:
		return uint(C.g_value_get_uint(v)), nil
	case C.G_TYPE_STRING:
		return C.GoString(C.g_value_get_string(v)), nil
	default:
		return nil, fmt.Errorf("Unsupported GValue type %d", t)
	}
}

// SetProperty sets property of the element.
func (s *Element) SetProperty(name string, val interface{}) error {
	v := C.newGValue()
	defer C.free(unsafe.Pointer(v))

	switch val := val.(type) {
	case int:
		C.g_value_init(v, C.G_TYPE_INT)
		C.g_value_set_int(v, C.int(val))
	case uint:
		C.g_value_init(v, C.G_TYPE_UINT)
		C.g_value_set_uint(v, C.uint(val))
	case string:
		cValue := C.CString(val)
		defer C.free(unsafe.Pointer(cValue))
		C.g_value_init(v, C.G_TYPE_STRING)
		C.g_value_set_string(v, cValue)
	default:
		return fmt.Errorf("Unsupported GValue type %d", reflect.TypeOf(val).Kind())
	}
	defer C.g_value_unset(v)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.setProperty(s.UnsafePointer(), cName, v)
	return nil
}
