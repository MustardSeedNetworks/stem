/*
 * Thread-safe errno formatting.
 *
 * strerror() returns a pointer into a buffer shared by every thread in the
 * process, so two threads formatting an error at once can each read the
 * other's message. The reflector and the dataplane are both multi-threaded,
 * which is why concurrency-mt-unsafe flags every use.
 */

#ifndef STEM_ERRNO_H
#define STEM_ERRNO_H

#include <string.h>

/*
 * Drop-in replacement for strerror(). The returned pointer is valid until the
 * next call *on the same thread*, which is the same contract callers already
 * relied on from strerror() — every call site formats the message immediately.
 */
static inline const char *stem_strerror(int errnum)
{
    static _Thread_local char buf[128];

#if defined(__GLIBC__)
    /*
     * glibc with _GNU_SOURCE (set for every target in the Makefile) provides
     * the GNU variant: it returns the message and may leave buf untouched when
     * it can hand back a static string, so the return value must be used
     * rather than buf.
     */
    return strerror_r(errnum, buf, sizeof buf);
#else
    /* POSIX/XSI variant (macOS, musl): fills buf and returns 0 on success. */
    if (strerror_r(errnum, buf, sizeof buf) != 0) {
        return "unknown error";
    }
    return buf;
#endif
}

#endif /* STEM_ERRNO_H */
