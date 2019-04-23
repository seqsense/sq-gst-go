/*
 * Copyright(c) 2019, SEQSENSE, Inc.
 * All rights reserved.
 */

/**
  \author Atsushi Watanabe (SEQSENSE, Inc.)
 **/

#ifndef APPSRC_H
#define APPSRC_H

#include <stdlib.h>
#include <gst/gst.h>
#include <gst/app/app.h>

void pushBuffer(void* element, void* buffer, int len);
GstState getState(void* element);

#endif  // APPSRC_H
