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

void pushBuffer(void* element, void* buffer, int len)
{
  GstBuffer* buffer_gst =
      gst_buffer_new_wrapped(g_memdup(buffer, len), len);
  gst_app_src_push_buffer(GST_APP_SRC(element), buffer_gst);
}

void sendEOS(void* element)
{
  gst_app_src_end_of_stream(GST_APP_SRC(element));
}
