/* (c) 2019 Confio UO. Licensed under Apache-2.0 */

/* Generated with cbindgen:0.9.1 */

/* Warning, this file is autogenerated by cbindgen. Don't modify this manually. */

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct Buffer {
  uint8_t *ptr;
  uintptr_t size;
} Buffer;

int32_t add(int32_t a, int32_t b);

void free_rust(Buffer buf);

Buffer greet(Buffer name);
