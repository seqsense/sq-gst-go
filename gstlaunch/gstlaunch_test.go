package gstlaunch

import (
	"reflect"
	"sort"
	"sync"
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
	l := MustNew("appsrc ! watchdog name=wd timeout=150 ! fakesink")

	errCh := make(chan struct{})
	l.RegisterErrorCallback(func(l *GstLaunch, e *gst.Element, msg string, dbgInfo string) {
		name, err := e.GetProperty("name")
		if err != nil {
			t.Errorf("failed to get name of the error source: %v", err)
		} else {
			if nameStr, ok := name.(string); !ok {
				t.Error("name of the element is not string")
			} else if nameStr != "wd" {
				t.Errorf("unexpected error source %s, expected \"wd\"", nameStr)
			}
		}
		if msg != "Watchdog triggered" {
			t.Errorf("unexpected error message %s, expected \"Watchdog triggered\"", msg)
		}
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

func TestGetAllElements(t *testing.T) {
	l := MustNew("audiotestsrc name=e0 ! queue name=e1 ! queue name=e2 ! fakesink name=e3")
	defer l.Kill()

	e, err := l.GetAllElements()
	if err != nil {
		t.Errorf("GetAllElement for active pipeline must not return error")
	}
	if len(e) != 4 {
		t.Fatalf("Unexpected number of the returned elements %d, expected %d", len(e), 4)
	}

	namesExpected := []string{"e0", "e1", "e2", "e3"}
	var names []string
	for i := 0; i < 4; i++ {
		name, err := e[i].GetProperty("name")
		if err != nil {
			t.Errorf("Failed to get name of element")
		}
		names = append(names, name.(string))
	}
	sort.Strings(names)
	if !reflect.DeepEqual(namesExpected, names) {
		t.Errorf("Unexpected names of the elements\ngot: %v\nexpected: %v", names, namesExpected)
	}
}

func TestKill(t *testing.T) {
	var wg sync.WaitGroup

	// Test segmentation fault of glib mainloop related race condition
	for i := 0; i < 5; i++ {
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				l, err := New("audiotestsrc ! queue ! fakesink")
				if err != nil {
					t.Errorf("Failed to create pipeline: %v", err)
					return
				}
				l.Start()
				l.Kill()
			}()
		}
		wg.Wait()
		// Wait for file descriptors releaesd to avoid fd limit
		i := 200
		for getNumCtx() > 0 {
			time.Sleep(10 * time.Millisecond)
			if i--; i <= 0 {
				t.Fatalf("Pipeline context is not cleared (%d remains)", getNumCtx())
			}
		}
	}
}
