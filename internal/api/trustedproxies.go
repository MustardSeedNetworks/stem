// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// trustedProxiesFromEnv reads STEM_TRUSTED_PROXIES, the CIDRs whose
// X-Forwarded-For may name the client that the auth rate limiter and the
// failed-login tracker count against (#962).
//
// Unset — the default — leaves loopback as the only trusted hop, exactly the
// behaviour #807 shipped, so the setting is additive and cannot loosen a
// deployment that does not use it. A malformed or all-trusting list is a
// startup failure rather than a setting that is silently ignored: an operator
// who believes their proxy is trusted and is wrong gets one shared bucket for
// every client, which is the awkwardness this exists to remove.
func trustedProxiesFromEnv() ([]netip.Prefix, error) {
	trustedProxies, err := logging.ParseTrustedProxies(os.Getenv(logging.TrustedProxiesEnv))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logging.TrustedProxiesEnv, err)
	}
	if len(trustedProxies) > 0 {
		logging.Info("Trusted proxies configured — forwarding headers from these peers "+
			"key rate limits and failed-login counters",
			"cidrs", trustedProxies, "event", "security.trusted_proxies")
	}
	return trustedProxies, nil
}
