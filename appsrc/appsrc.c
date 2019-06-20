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

void pushBuffer(void* element, void* buffer, int len)
{
  GstBuffer* buffer_gst =
      gst_buffer_new_wrapped_full(0, buffer, len, 0, len, 0, 0);
  gst_app_src_push_buffer(GST_APP_SRC(element), buffer_gst);
}
