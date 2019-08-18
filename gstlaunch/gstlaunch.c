/*
 * Copyright(c) 2019, SEQSENSE, Inc.
 * All rights reserved.
 */

/**
  \author Atsushi Watanabe (SEQSENSE, Inc.)
 **/

#include <stdio.h>
#include <stdlib.h>

#include <gst/gst.h>

#include "gstlaunch.h"

static GMutex g_mutex;
static GMainLoop* g_mainloop;

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

  g_mutex_lock(&ctx->mutex);
  if (ctx->closed >= CLOSING)
  {
    if (ctx->closed == CLOSED)
      fprintf(stderr, "Received message from removed source: %d\n", ctx->user_int);
    ctx->closed = CLOSED;
    g_mutex_unlock(&ctx->mutex);
    return FALSE;
  }
  g_mutex_unlock(&ctx->mutex);

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_EOS))
    goCbEOS(ctx->user_int);

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_ERROR))
    goCbError(ctx->user_int);

  if ((GST_MESSAGE_TYPE(msg) & GST_MESSAGE_STATE_CHANGED))
  {
    if ((void*)GST_MESSAGE_SRC(msg) == (void*)ctx->pipeline)
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

  g_mutex_lock(&g_mutex);
  pipeline = gst_parse_launch(launch, &err);
  if (pipeline == NULL)
  {
    g_mutex_unlock(&g_mutex);
    fprintf(stderr, "gst_parse_launch failed: %s\n", err->message);
    return NULL;
  }
  ctx = malloc(sizeof(Context));
  if (ctx == NULL)
  {
    g_mutex_unlock(&g_mutex);
    fprintf(stderr, "failed to allocate memory for gstlaunch context\n");
    return NULL;
  }
  ctx->pipeline = pipeline;
  ctx->user_int = user_int;
  ctx->closed = IDLE;
  g_mutex_init(&ctx->mutex);

  GstBus* bus = gst_element_get_bus(ctx->pipeline);
  ctx->watch_tag = gst_bus_add_watch(bus, cbMessage, ctx);
  g_object_unref(bus);

  g_mutex_unlock(&g_mutex);
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
  g_mutex_lock(&ctx->mutex);
  ctx->closed = CLOSING;
  g_mutex_unlock(&ctx->mutex);

  gst_element_set_state(ctx->pipeline, GST_STATE_NULL);
  gst_object_unref(ctx->pipeline);
}
void pipelineFree(Context* ctx)
{
  g_mutex_lock(&g_mutex);
  g_mutex_lock(&ctx->mutex);
  if (ctx->closed == CLOSING)
  {
    ctx->closed = CLOSED;
    g_source_remove(ctx->watch_tag);
  }
  g_mutex_unlock(&ctx->mutex);

  free(ctx);
  g_mutex_unlock(&g_mutex);
}
GstElement* getElement(Context* ctx, const char* name)
{
  return gst_bin_get_by_name(GST_BIN(ctx->pipeline), name);
}
