/*
 * Payload signatures the Stem test master writes at UDP payload offset 0.
 *
 * The test master (rfc2544.h) stamps its frames with these and the reflector
 * (reflector.h) reflects every one of them. Both include this file, so a test
 * stream cannot be added without the reflector learning its signature: one
 * that was missing got no frames back and reported 100 % loss (#1531).
 *
 * Signatures shorter than their field are space-padded on the wire.
 */

#ifndef STEM_SIGNATURES_H
#define STEM_SIGNATURES_H

// NOLINTBEGIN(cppcoreguidelines-macro-to-enum,modernize-macro-to-enum)
#define RFC2544_SIGNATURE    "RFC254"
#define RFC2544_SIG_LEN      6
#define Y1564_SIGNATURE      "Y.1564 "
#define Y1564_SIG_LEN        7
#define Y1731_SIGNATURE      "Y.1731 "
#define Y1731_SIG_LEN        7
#define MEF_SIGNATURE        "MEF48 "
#define MEF_SIG_LEN          7
#define TSN_SIGNATURE        "802Qbv"
#define TSN_SIG_LEN          7
#define TRAFFICGEN_SIGNATURE "CUSTOM "
// NOLINTEND(cppcoreguidelines-macro-to-enum,modernize-macro-to-enum)

/* Every signature above, for code that must accept all of them. */
#define STEM_TEST_SIGNATURES(X) \
    X(RFC2544_SIGNATURE)        \
    X(Y1564_SIGNATURE)          \
    X(Y1731_SIGNATURE)          \
    X(MEF_SIGNATURE)            \
    X(TSN_SIGNATURE)            \
    X(TRAFFICGEN_SIGNATURE)

#endif /* STEM_SIGNATURES_H */
