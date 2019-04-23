package gst

import (
	"testing"

	"github.com/seqsense/sq-gst-go/internal/dummyelement"
)

func TestGstLaunch(t *testing.T) {
	e := NewGstElement(dummyelement.New())
	s0 := e.State()
	if s0 != GST_STATE_NULL {
		t.Errorf("Element state at initialization must be GST_STATE_NULL(%d) but got %d", GST_STATE_NULL, s0)
	}
}
