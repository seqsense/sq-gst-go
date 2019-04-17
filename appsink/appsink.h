/*
 * Copyright(c) 2019, SEQSENSE, Inc.
 * All rights reserved.
 */

/**
  \author Atsushi Watanabe (SEQSENSE, Inc.)
 **/

#ifndef APPSINK_H
#define APPSINK_H

#include <stdlib.h>
#include <gst/gst.h>
#include <gst/app/app.h>

extern void goBufferHandler(void* buffer, int len, int samples, int id);

typedef struct
{
  int id;
} HandlerUserData;

void registerBufferHandler(void* element, int id);

#endif  // APPSINK_H
