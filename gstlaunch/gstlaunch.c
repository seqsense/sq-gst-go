/*
 * Copyright(c) 2019, SEQSENSE, Inc.
 * All rights reserved.
 */

/**
  \author Atsushi Watanabe (SEQSENSE, Inc.)
 **/

#include <stdlib.h>
#include <gst/gst.h>

#include "gstlaunch.h"

void init()
{
  int argc = 1;
  char* exec_name = "rtsp_receiver";
  char** argv = &exec_name;
  gst_init(&argc, &argv);
}
static gboolean cbMessage(GstBus* bus, GstMessage* msg, gpointer p)
{
  Context* ctx = (Context*)p;

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_EOS))
    goCbEOS(ctx->user_int);

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_ERROR))
    goCbError(ctx->user_int);

  return TRUE;
}
Context* create(const char* launch, int user_int)
{
  Context* ctx;
  GMainLoop* mainloop;
  GstElement* pipeline;
  GError* err = NULL;
  GstBus* bus;
  GstElement* src;

  mainloop = g_main_loop_new(NULL, FALSE);

  pipeline = gst_parse_launch(launch, &err);
  if (err != NULL)
  {
    g_object_unref(mainloop);
    return NULL;
  }
  ctx = malloc(sizeof(Context));
  ctx->mainloop = mainloop;
  ctx->pipeline = pipeline;
  ctx->user_int = user_int;

  bus = gst_element_get_bus(pipeline);
  gst_bus_add_watch(bus, cbMessage, ctx);
  g_object_unref(bus);

  return ctx;
}
void mainloopRun(Context* ctx)
{
  gst_element_set_state(ctx->pipeline, GST_STATE_PLAYING);
  g_main_loop_run(ctx->mainloop);
}
void mainloopKill(Context* ctx)
{
  gst_element_set_state(ctx->pipeline, GST_STATE_NULL);
  gst_object_unref(ctx->pipeline);
  g_main_loop_quit(ctx->mainloop);
}
GstElement* getElement(Context* ctx, const char* name)
{
  return gst_bin_get_by_name(GST_BIN(ctx->pipeline), name);
}
