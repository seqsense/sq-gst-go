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

GMainLoop* g_mainloop;

void init(char* exec_name)
{
  int argc = 1;
  char** argv = &exec_name;
  gst_init(&argc, &argv);

  g_mainloop = g_main_loop_new(NULL, FALSE);
}
void runMainloop()
{
  g_main_loop_run(g_mainloop);
}

static gboolean cbMessage(GstBus* bus, GstMessage* msg, gpointer p)
{
  Context* ctx = (Context*)p;

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_EOS))
    goCbEOS(ctx->user_int);

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_ERROR))
    goCbError(ctx->user_int);

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_STATE_CHANGED))
  {
    if (GST_MESSAGE_SRC(msg) == GST_OBJECT(ctx->pipeline))
    {
      GstState old_state, new_state, pending_state;
      gst_message_parse_state_changed(msg, &old_state, &new_state, &pending_state);
      goCbState(ctx->user_int, old_state, new_state, pending_state);
    }
  }

  return TRUE;
}
Context* create(const char* launch, int user_int)
{
  Context* ctx;
  GstElement* pipeline;
  GError* err = NULL;
  GstElement* src;

  pipeline = gst_parse_launch(launch, &err);
  if (err != NULL)
  {
    return NULL;
  }
  ctx = malloc(sizeof(Context));
  ctx->pipeline = pipeline;
  ctx->user_int = user_int;

  ctx->bus = gst_element_get_bus(pipeline);
  ctx->watch_tag = gst_bus_add_watch(ctx->bus, cbMessage, ctx);

  return ctx;
}
void pipelineStart(Context* ctx)
{
  gst_element_set_state(ctx->pipeline, GST_STATE_PLAYING);
}
void pipelineStop(Context* ctx)
{
  gst_element_set_state(ctx->pipeline, GST_STATE_NULL);
}
void pipelineUnref(Context* ctx)
{
  gst_element_set_state(ctx->pipeline, GST_STATE_NULL);
  g_object_unref(ctx->bus);
  g_source_remove(ctx->watch_tag);
  gst_object_unref(ctx->pipeline);
  free(ctx);
}
GstElement* getElement(Context* ctx, const char* name)
{
  return gst_bin_get_by_name(GST_BIN(ctx->pipeline), name);
}
