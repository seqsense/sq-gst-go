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

inline void unrefElement(void* element)
{
  gst_object_unref(element);
}

void pushBuffer(void* element, void* buffer, int len);

#endif  // APPSRC_H
