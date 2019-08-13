package gst

import (
	"reflect"
	"testing"

	"github.com/seqsense/sq-gst-go/internal/dummyelement"
)

func TestNewElement(t *testing.T) {
	e := NewElement(dummyelement.New())
	s0 := e.State()
	if s0 != StateNull {
		t.Errorf("Element state at initialization must be StateNull(%d) but got %d", StateNull, s0)
	}
}

func TestGetProperty_Int(t *testing.T) {
	e := NewElement(dummyelement.New())
	p, err := e.GetProperty("num-buffers")
	if err != nil {
		t.Fatalf("Failed to GetProperty: %v", err)
	}
	switch v := p.(type) {
	case int:
		if v != -1 {
			t.Errorf("fakesink.num-buffers must be -1, but got %d", v)
		}
	default:
		t.Fatalf("Wrong return value type: %s", reflect.TypeOf(v).Kind())
	}
}

func TestSetProperty_Int(t *testing.T) {
	e := NewElement(dummyelement.New())
	if err := e.SetProperty("num-buffers", 11); err != nil {
		t.Fatalf("Failed to SetProperty: %v", err)
	}

	p, err := e.GetProperty("num-buffers")
	if err != nil {
		t.Fatalf("Failed to GetProperty: %v", err)
	}
	switch v := p.(type) {
	case int:
		if v != 11 {
			t.Errorf("fakesink.num-buffers must be 11, but got %d", v)
		}
	default:
		t.Fatalf("Wrong return value type: %s", reflect.TypeOf(v).Kind())
	}
}

func TestSetProperty_String(t *testing.T) {
	e := NewElement(dummyelement.New())
	if err := e.SetProperty("name", "the-element"); err != nil {
		t.Fatalf("Failed to SetProperty: %v", err)
	}

	p, err := e.GetProperty("name")
	if err != nil {
		t.Fatalf("Failed to GetProperty: %v", err)
	}
	switch v := p.(type) {
	case string:
		if v != "the-element" {
			t.Errorf("fakesink.name must be \"the-element\", but got %s", v)
		}
	default:
		t.Fatalf("Wrong return value type: %s", reflect.TypeOf(v).Kind())
	}
}
