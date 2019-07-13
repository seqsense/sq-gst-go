package gstlaunch

import (
	"context"
	"fmt"
	"testing"
	"time"

	gst "github.com/seqsense/sq-gst-go"
	"github.com/seqsense/sq-gst-go/appsrc"
)

func TestGstLaunch(t *testing.T) {
	startMethod := map[string]func(l *GstLaunch){
		"GstLaunch.Run": func(l *GstLaunch) {
			go l.Run(context.Background())
		},
		"GstLaunch.Start": func(l *GstLaunch) {
			l.Start()
		},
	}
	for name, start := range startMethod {
		t.Run(name, func(t *testing.T) {
			l := New("audiotestsrc ! queue ! fakesink")
			defer l.Unref()

			if l.Active() != false {
				t.Errorf("pipeline must be inactive before Run()")
			}

			start(l)

			<-time.After(time.Millisecond * 100)
			if l.Active() != true {
				t.Errorf("pipeline must be active after Run()")
			}

			l.Kill()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if l.Wait(ctx) != nil {
				t.Errorf("failed to wait pipeline stop")
			}

			if l.Active() != false {
				t.Errorf("pipeline must be inactive after Kill()")
			}
		})
	}
}

func TestGstLaunch_eosHandling(t *testing.T) {
	eosCh := make(chan struct{})

	l := New("appsrc name=src ! watchdog timeout=150 ! fakesink")
	defer l.Unref()
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
		t.Errorf("expected error message, but timed-out")
	case <-eosCh:
	}

	l.Kill()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if l.Wait(ctx) != nil {
		t.Errorf("failed to wait pipeline stop")
	}
}

func TestGstLaunch_errorHandling(t *testing.T) {
	errCh := make(chan struct{})

	l := New("appsrc ! watchdog timeout=150 ! fakesink")
	defer l.Unref()
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if l.Wait(ctx) != nil {
		t.Errorf("failed to wait pipeline stop")
	}
}

func TestGetElement(t *testing.T) {
	l := New("audiotestsrc ! queue name=named_elem ! queue ! fakesink")
	defer l.Unref()

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

	l.Start()
	<-time.After(time.Millisecond * 100)

	if s := e.State(); s != gst.GST_STATE_PLAYING {
		t.Errorf("GstElement state must be GST_STATE_PLAYING(%d) after Run(), but got %d", gst.GST_STATE_PLAYING, s)
	}

	l.Kill()
}
