package gstlaunch

import (
	"testing"
	"time"

	gst "github.com/seqsense/sq-gst-go"
	"github.com/seqsense/sq-gst-go/appsrc"
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

func TestGstLaunch_eosHandling(t *testing.T) {
	eosCh := make(chan struct{})

	l := New("appsrc name=src ! watchdog timeout=150 ! fakesink")
	l.RegisterEOSCallback(func(l *GstLaunch) {
		eosCh <- struct{}{}
	})
	srcElem, err := l.GetElement("src")
	if err != nil {
		t.Fatalf("failed to get appsrc element: %v", err)
	}
	src := appsrc.New(srcElem)

	go func() {
		l.Run()
	}()

	select {
	case <-time.After(time.Millisecond * 100):
	case <-eosCh:
		t.Errorf("unexpected EOS message")
	}
	src.EOS()
	select {
	case <-time.After(time.Millisecond * 100):
		t.Errorf("expected error message, but timed-out")
	case <-eosCh:
	}

	l.Kill()
	l.Wait()
}

func TestGstLaunch_errorHandling(t *testing.T) {
	errCh := make(chan struct{})

	l := New("appsrc ! watchdog timeout=150 ! fakesink")
	l.RegisterErrorCallback(func(l *GstLaunch) {
		errCh <- struct{}{}
	})
	go func() {
		l.Run()
	}()

	select {
	case <-time.After(time.Millisecond * 100):
	case <-errCh:
		t.Errorf("unexpected error message")
	}
	select {
	case <-time.After(time.Millisecond * 100):
		t.Errorf("expected error message, but timed-out")
	case <-errCh:
	}

	l.Kill()
	l.Wait()
}

func TestGetElement(t *testing.T) {
	l := New("audiotestsrc ! queue name=named_elem ! queue ! fakesink")

	e, err := l.GetElement("named_elem")
	if err != nil {
		t.Errorf("GstElement for existing element must not return error")
	}
	if e == nil {
		t.Errorf("GstElement for existing element must return pointer")
	}
	if s := e.State(); s != gst.GST_STATE_NULL {
		t.Errorf("GstElement state must be GST_STATE_NULL(%d) at initialization, but got %d", gst.GST_STATE_NULL, s)
	}

	e_inexistent, err := l.GetElement("inexistent_elem")
	if err == nil {
		t.Errorf("GstElement for inexistent element must return error")
	}
	if e_inexistent != nil {
		t.Errorf("GstElement for inexistent element must return nil pointer")
	}

	go func() {
		l.Run()
	}()
	<-time.After(time.Millisecond * 100)

	if s := e.State(); s != gst.GST_STATE_PLAYING {
		t.Errorf("GstElement state must be GST_STATE_PLAYING(%d) after Run(), but got %d", gst.GST_STATE_PLAYING, s)
	}

	l.Kill()
}
