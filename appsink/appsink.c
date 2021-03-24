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
      GstMapInfo info;
      if (gst_buffer_map(buffer, &info, GST_MAP_READ))
      {
        goBufferHandler(info.data, info.size, GST_BUFFER_DURATION(buffer), ud->id);
        gst_buffer_unmap(buffer, &info);
      }
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
