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

#ifndef GSTLAUNCH_H
#define GSTLAUNCH_H

#include <stdlib.h>
#include <gst/gst.h>

typedef struct
{
  GMutex mutex;
  GstElement* pipeline;
  int user_int;
  unsigned int watch_tag;
  enum
  {
    IDLE,
    CLOSING,
    CLOSED,
  } closed;
} Context;

extern void goCbEOS(int id);
extern void goCbError(
    int id, void* src, char* msg, int msg_size, char* dbg_info, int dbg_info_size);
extern void goCbState(
    int id, unsigned int old_state, unsigned int new_state, unsigned int pending_state);

void init(char* exec_name);
Context* create(const char* launch, int user_int);
void pipelineStart(Context* ctx);
void pipelineStop(Context* ctx);
void pipelineUnref(Context* ctx);
void pipelineFree(Context* ctx);
GstElement* getElement(Context* ctx, const char* name);
GstElement** getAllElements(Context* ctx);
GstElement* elementAt(GstElement** es, const int i);
void refElement(void* e);

#endif  // GSTLAUNCH_H
