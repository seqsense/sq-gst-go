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
} Context;

extern void goCbEOS(int id);
extern void goCbError(int id);

void init(char* exec_name);
void runMainloop();
Context* create(const char* launch, int user_int);
void pipelineStart(Context* ctx);
void pipelineKill(Context* ctx);
GstElement* getElement(Context* ctx, const char* name);

#endif  // GSTLAUNCH_H
