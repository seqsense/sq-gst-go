package gst

import (
	"testing"

	"github.com/seqsense/sq-gst-go/internal/dummyelement"
)

func TestLaunch(t *testing.T) {
	e := NewElement(dummyelement.New())
	s0 := e.State()
	if s0 != StateNull {
		t.Errorf("Element state at initialization must be StateNull(%d) but got %d", StateNull, s0)
	}
}
