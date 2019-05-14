/*
 * Copyright(c) 2019, SEQSENSE, Inc.
 * All rights reserved.
 */

/**
  \author Atsushi Watanabe (SEQSENSE, Inc.)
 **/

#include <stdlib.h>
#include <gst/gst.h>
#include <gst/app/app.h>

#include "appsink.h"

GstFlowReturn bufferHandlerC(GstElement* element, gpointer user_data)
{
  HandlerUserData* ud = (HandlerUserData*)user_data;

  GstSample* sample = NULL;
  g_signal_emit_by_name(element, "pull-sample", &sample);
  if (sample)
  {
    GstBuffer* buffer = gst_sample_get_buffer(sample);
    if (buffer)
    {
      gpointer copy = NULL;
      gsize size = 0;
      gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &size);
      goBufferHandler(copy, size, GST_BUFFER_DURATION(buffer), ud->id);
      g_free(copy);
    }
    gst_sample_unref(sample);
  }

  return GST_FLOW_OK;
}

void registerBufferHandler(void* element, int id)
{
  HandlerUserData* ud = (HandlerUserData*)malloc(sizeof(HandlerUserData));
  ud->id = id;

  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(bufferHandlerC), ud);
}
