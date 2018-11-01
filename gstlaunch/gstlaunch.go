package gstlaunch

import (
	"unsafe"
)

// #cgo pkg-config: gobject-2.0 gstreamer-1.0 gstreamer-base-1.0 gstreamer-rtsp-server-1.0 gstreamer-rtsp-1.0
// #include <stdlib.h>
// #include <stdio.h>
// #include <gst/gst.h>
//
// typedef void* Mainloop;
//
// void init()
// {
//   int argc = 1;
//   char *exec_name = "rtsp_receiver";
//   char **argv = &exec_name;
//   gst_init(&argc, &argv);
// }
// gboolean cbMessage(GstBus *bus, GstMessage *msg, gpointer p)
// {
//   GMainLoop *mainloop = (GMainLoop*)p;
//
//   if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_EOS))
//     g_main_loop_quit(mainloop);
//
//   if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_ERROR))
//     g_main_loop_quit(mainloop);
//
//   return TRUE;
// }
// Mainloop create(const char *launch)
// {
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
//   bus = gst_element_get_bus(pipeline);
//   gst_bus_add_watch(bus, cbMessage, mainloop);
//   g_object_unref(bus);
//
//   gst_element_set_state(pipeline, GST_STATE_PLAYING);
//
//   return mainloop;
// }
// void mainloopRun(Mainloop p)
// {
//   GMainLoop *mainloop = (GMainLoop*)p;
//   g_main_loop_run(mainloop);
// }
// void mainloopKill(Mainloop p)
// {
//   GMainLoop *mainloop = (GMainLoop*)p;
//   g_main_loop_quit(mainloop);
// }
import "C"

func init() {
	C.init()
}

type GstLaunch struct {
	mainloop C.Mainloop
	quit     chan bool
	active   bool
}

func New(launch string) *GstLaunch {
	c_launch := C.CString(launch)
	defer C.free(unsafe.Pointer(c_launch))

	mainloop := C.create(c_launch)
	if mainloop == nil {
		panic("Failed to parse gst-launch text")
	}

	return &GstLaunch{mainloop: mainloop, quit: make(chan bool, 1)}
}

func (s *GstLaunch) Run() {
	s.active = true
	C.mainloopRun(s.mainloop)
	s.quit <- true
	s.active = false
}

func (s *GstLaunch) Wait() {
	<-s.quit
}

func (s *GstLaunch) Kill() {
	C.mainloopKill(s.mainloop)
}

func (s *GstLaunch) Active() bool {
	if s == nil {
		return false
	}
	return s.active
}
