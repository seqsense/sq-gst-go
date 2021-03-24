// Copyright 2021 SEQSENSE, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appsink

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/seqsense/sq-gst-go/appsrc"
	"github.com/seqsense/sq-gst-go/gstlaunch"
)

func TestAppSrcAppSink(t *testing.T) {
	l := gstlaunch.MustNew("appsrc name=src ! appsink name=sink")

	var received []byte
	gstSink, err := l.GetElement("sink")
	if err != nil {
		t.Fatalf("appsink element must be got")
	}
	sink := New(gstSink, func(b []byte, samples int) {
		received = b
	})
	defer sink.Close()

	gstSrc, err := l.GetElement("src")
	if err != nil {
		t.Fatalf("appsrc element must be got")
	}
	src := appsrc.New(gstSrc)

	// Any used objects must not finalized
	runtime.GC()

	l.Start()

	<-time.After(time.Millisecond * 100)
	pushed := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	src.PushBuffer(pushed)
	<-time.After(time.Millisecond * 100)

	l.Kill()
	if len(received) == 0 {
		t.Errorf("appsink must receive a buffer")
	} else if bytes.Compare(received, pushed) != 0 {
		t.Errorf("appsink received wrong buffer, expected: %v, received: %v", pushed, received)
	}
}
