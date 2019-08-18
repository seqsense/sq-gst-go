/*
 * Copyright(c) 2019, SEQSENSE, Inc.
 * All rights reserved.
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
  GstElement* pipeline;
  int user_int;
  unsigned int watch_tag;
} Context;

extern void goCbEOS(int id);
extern void goCbError(int id);
extern void goCbState(
    int id, unsigned int old_state, unsigned int new_state, unsigned int pending_state);

void init(char* exec_name);
void runMainloop();
Context* create(const char* launch, int user_int);
void pipelineStart(Context* ctx);
void pipelineStop(Context* ctx);
void pipelineUnref(Context* ctx);
GstElement* getElement(Context* ctx, const char* name);

#endif  // GSTLAUNCH_H
