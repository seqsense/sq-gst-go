package gstlaunch

import (
	"fmt"
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0 gstreamer-1.0 gstreamer-base-1.0
// #include <stdlib.h>
// #include <gst/gst.h>
//
// typedef struct
// {
//   GMainLoop *mainloop;
//   GstElement *pipeline;
//   int user_int;
// } Context;
//
// extern void goCbEOS(int);
// extern void goCbError(int);
//
// static void init()
// {
//   int argc = 1;
//   char *exec_name = "sq-gst-go";
//   char **argv = &exec_name;
//   gst_init(&argc, &argv);
// }
// static gboolean cbMessage(GstBus *bus, GstMessage *msg, gpointer p)
// {
//   Context *ctx = (Context*)p;
//
//   if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_EOS))
//     goCbEOS(ctx->user_int);
//
//   if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_ERROR))
//     goCbError(ctx->user_int);
//
//   return TRUE;
// }
// static Context *create(const char *launch, int user_int)
// {
//   Context *ctx;
//   GMainLoop *mainloop;
//   GstElement *pipeline;
//   GError *err = NULL;
//   GstBus *bus;
//   GstElement *src;
//
//   mainloop = g_main_loop_new(NULL, FALSE);
//
//   pipeline = gst_parse_launch(launch, &err);
//   if (err != NULL)
//   {
//     g_object_unref(mainloop);
//     return NULL;
//   }
//   ctx = malloc(sizeof(Context));
//   ctx->mainloop = mainloop;
//   ctx->pipeline = pipeline;
//   ctx->user_int = user_int;
//
//   bus = gst_element_get_bus(pipeline);
//   gst_bus_add_watch(bus, cbMessage, ctx);
//   g_object_unref(bus);
//
//   gst_element_set_state(pipeline, GST_STATE_PLAYING);
//
//   return ctx;
// }
// static void mainloopRun(Context *ctx)
// {
//   g_main_loop_run(ctx->mainloop);
// }
// static void mainloopKill(Context *ctx)
// {
//   gst_element_set_state(ctx->pipeline, GST_STATE_NULL);
//   gst_object_unref(ctx->pipeline);
//   g_main_loop_quit(ctx->mainloop);
// }
import "C"

func init() {
	C.init()
}

type GstLaunch struct {
	ctx     *C.Context
	quit    chan bool
	active  bool
	cbEOS   func(*GstLaunch)
	cbError func(*GstLaunch)
}

var (
	cPointerMap          = make(map[int]*GstLaunch)
	cPointerMapIndex int = 0
)

func New(launch string) *GstLaunch {
	c_launch := C.CString(launch)
	defer C.free(unsafe.Pointer(c_launch))

	l := &GstLaunch{quit: make(chan bool, 1), active: false, cbEOS: nil, cbError: nil}
	cPointerMap[cPointerMapIndex] = l

	ctx := C.create(c_launch, C.int(cPointerMapIndex))
	if ctx == nil {
		panic("Failed to parse gst-launch text")
	}
	l.ctx = ctx

	cPointerMapIndex++

	runtime.SetFinalizer(l, finalizeGstLaunch)

	return l
}

func finalizeGstLaunch(s *GstLaunch) {
	C.free(unsafe.Pointer(s.ctx))
}

func (s *GstLaunch) RegisterErrorCallback(f func(*GstLaunch)) {
	s.cbError = f
}

func (s *GstLaunch) RegisterEOSCallback(f func(*GstLaunch)) {
	s.cbEOS = f
}

//export goCbEOS
func goCbEOS(i C.int) {
	s, ok := cPointerMap[int(i)]
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	if s.cbEOS != nil {
		s.cbEOS(s)
	}
}

//export goCbError
func goCbError(i C.int) {
	s, ok := cPointerMap[int(i)]
	if !ok {
		panic(fmt.Errorf("Failed to map pointer from cgo func (%d)", int(i)))
	}
	if s.cbError != nil {
		s.cbError(s)
	}
}

func (s *GstLaunch) Run() {
	s.active = true
	C.mainloopRun(s.ctx)
	s.quit <- true
	s.active = false
}

func (s *GstLaunch) Wait() {
	<-s.quit
}
func (s *GstLaunch) Kill() {
	C.mainloopKill(s.ctx)
}

func (s *GstLaunch) Active() bool {
	if s == nil {
		return false
	}
	return s.active
}
