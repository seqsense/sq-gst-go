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
