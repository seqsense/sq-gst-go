package gstlaunch

import (
	"testing"
	"time"
)

func TestGstLaunch(t *testing.T) {
	l := New("audiotestsrc ! queue ! fakesink")

	if l.Active() != false {
		t.Errorf("pipeline must be inactive before Run()")
	}

	go func() {
		l.Run()
	}()

	<-time.After(time.Millisecond * 100)
	if l.Active() != true {
		t.Errorf("pipeline must be active after Run()")
	}

	l.Kill()
	l.Wait()

	if l.Active() != false {
		t.Errorf("pipeline must be inactive after Kill()")
	}
}

func TestGetElement(t *testing.T) {
	l := New("audiotestsrc ! queue name=named_elem ! queue ! fakesink")

	go func() {
		l.Run()
	}()
	<-time.After(time.Millisecond * 100)

	e, err := l.GetElement("named_elem")
	if err != nil {
		t.Errorf("GstElement for existing element must not return error")
	}
	if e == nil {
		t.Errorf("GstElement for existing element must return pointer")
	}

	e_inexistent, err := l.GetElement("inexistent_elem")
	if err == nil {
		t.Errorf("GstElement for inexistent element must return error")
	}
	if e_inexistent != nil {
		t.Errorf("GstElement for inexistent element must return nil pointer")
	}

	l.Kill()
}
