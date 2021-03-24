/* Copyright 2021 SEQSENSE, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
  \author Atsushi Watanabe (SEQSENSE, Inc.)
 **/

#include <stdio.h>
#include <stdlib.h>

#include <gst/gst.h>

#include "gstlaunch.h"

static GMutex g_mutex;

gpointer runMainloop(gpointer mainloop)
{
  g_main_loop_run(mainloop);
  return NULL;
}
void init(char* exec_name)
{
  int argc = 1;
  char** argv = &exec_name;
  gst_init(&argc, &argv);

  GMainLoop* mainloop = g_main_loop_new(NULL, FALSE);
  GThread* thread = g_thread_new("mainloop", runMainloop, mainloop);
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
  {
    GError* err = NULL;
    gchar* dbg_info = NULL;

    gst_message_parse_error(msg, &err, &dbg_info);
    int dbg_info_size = 0;
    if (dbg_info != NULL)
      dbg_info_size = strlen(dbg_info);

    goCbError(
        ctx->user_int, (void*)GST_MESSAGE_SRC(msg),
        err->message, strlen(err->message), dbg_info, dbg_info_size);

    g_error_free(err);
    g_free(dbg_info);
  }

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
    gst_object_unref(ctx->pipeline);
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

  if (ctx->watch_tag == 0)
  {
    fprintf(stderr, "failed to add watch to gstlaunch context\n");
    gst_object_unref(ctx->pipeline);
    free(ctx);
    g_mutex_unlock(&g_mutex);
    return NULL;
  }

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
  gst_object_unref(ctx->pipeline);
  g_mutex_unlock(&ctx->mutex);

  g_mutex_clear(&ctx->mutex);
  free(ctx);
  g_mutex_unlock(&g_mutex);
}
GstElement* getElement(Context* ctx, const char* name)
{
  return gst_bin_get_by_name(GST_BIN(ctx->pipeline), name);
}
GstElement** getAllElements(Context* ctx)
{
  GstElement** elements =
      malloc(sizeof(GstElement*) * (GST_BIN_NUMCHILDREN(GST_BIN(ctx->pipeline)) + 1));
  int i = 0;
  GstIterator* it = gst_bin_iterate_elements(GST_BIN(ctx->pipeline));

  for (gboolean done = FALSE; !done;)
  {
    GValue val = G_VALUE_INIT;
    switch (gst_iterator_next(it, &val))
    {
      case GST_ITERATOR_OK:
      {
        elements[i] = g_value_get_object(&val);
        gst_object_ref(elements[i]);
        ++i;
        g_value_unset(&val);
        break;
      }
      default:
      {
        done = TRUE;
        break;
      }
    }
  }
  gst_iterator_free(it);
  elements[i] = NULL;
  return elements;
}
GstElement* elementAt(GstElement** es, const int i)
{
  return es[i];
}
void refElement(void* e)
{
  gst_object_ref(GST_ELEMENT(e));
}
