package gstlaunch

import (
	"testing"
	"time"

	gst "github.com/seqsense/sq-gst-go"
	"github.com/seqsense/sq-gst-go/appsrc"
)

func TestLaunch(t *testing.T) {
	l := MustNew("audiotestsrc ! queue ! fakesink")

	if l.Active() != false {
		t.Errorf("pipeline must be inactive before Start()")
	}

	l.Start()

	<-time.After(time.Millisecond * 100)
	if l.Active() != true {
		t.Errorf("pipeline must be active after Start()")
	}

	l.Kill()

	<-time.After(time.Millisecond * 100)
	if l.Active() != false {
		t.Errorf("pipeline must be inactive after Kill()")
	}
}

func TestLaunch_eosHandling(t *testing.T) {
	l := MustNew("appsrc name=src ! fakesink")

	eosCh := make(chan struct{})
	l.RegisterEOSCallback(func(l *GstLaunch) {
		eosCh <- struct{}{}
	})
	srcElem, err := l.GetElement("src")
	if err != nil {
		t.Fatalf("failed to get appsrc element: %v", err)
	}
	src := appsrc.New(srcElem)

	l.Start()

	select {
	case <-time.After(time.Millisecond * 100):
	case <-eosCh:
		t.Errorf("unexpected EOS message")
	}
	src.EOS()
	select {
	case <-time.After(time.Millisecond * 100):
		t.Errorf("expected EOS message, but timed-out")
	case <-eosCh:
	}

	l.Kill()
}

func TestLaunch_errorHandling(t *testing.T) {
	l := MustNew("appsrc ! watchdog timeout=150 ! fakesink")

	errCh := make(chan struct{})
	l.RegisterErrorCallback(func(l *GstLaunch) {
		errCh <- struct{}{}
	})
	l.Start()

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
}

func TestLaunch_stateHandling(t *testing.T) {
	l := MustNew("audiotestsrc ! queue ! fakesink")

	stateCh := make(chan gst.State, 100)
	l.RegisterStateCallback(func(l *GstLaunch, _, s, _ gst.State) {
		stateCh <- s
	})

	l.Start()
L1:
	for {
		select {
		case <-time.After(time.Millisecond * 100):
			t.Error("expected state callback, but timed-out")
			break L1
		case s := <-stateCh:
			if s == gst.StatePlaying {
				break L1
			}
		}
	}

	l.Kill()
L2:
	for {
		select {
		case <-time.After(time.Millisecond * 100):
			t.Error("expected state callback, but timed-out")
			break L2
		case s := <-stateCh:
			if s == gst.StateNull {
				break L2
			}
		}
	}
}

func TestGetElement(t *testing.T) {
	l := MustNew("audiotestsrc ! queue name=named_elem ! queue ! fakesink")

	e, err := l.GetElement("named_elem")
	if err != nil {
		t.Errorf("Element for existing element must not return error")
	}
	if e == nil {
		t.Errorf("Element for existing element must return pointer")
	}
	if s := e.State(); s != gst.StateNull {
		t.Errorf("Element state must be StateNull(%d) at initialization, but got %d", gst.StateNull, s)
	}

	eInexistent, err := l.GetElement("inexistent_elem")
	if err == nil {
		t.Errorf("Element for inexistent element must return error")
	}
	if eInexistent != nil {
		t.Errorf("Element for inexistent element must return nil pointer")
	}

	l.Start()
	<-time.After(time.Millisecond * 100)

	if s := e.State(); s != gst.StatePlaying {
		t.Errorf("Element state must be StatePlaying(%d) after Start(), but got %d", gst.StatePlaying, s)
	}

	l.Kill()
}
