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
  GMainLoop* mainloop;
  GstElement* pipeline;
  int user_int;
} Context;

extern void goCbEOS(int id);
extern void goCbError(int id);

void init();
Context* create(const char* launch, int user_int);
void mainloopRun(Context* ctx);
void mainloopKill(Context* ctx);

#endif  // GSTLAUNCH_H
