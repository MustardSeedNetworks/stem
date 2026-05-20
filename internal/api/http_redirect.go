// SPDX-License-Identifier: BUSL-1.1

package api

// http_redirect.go provides the tiny HTTP listener that 308-redirects
// every request to the HTTPS port. Adapted from seed's
// internal/api/server_lifecycle.go (httpToHTTPSRedirectHandler /
// startHTTPRedirect); keep in sync.

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/krisarmstrong/stem/internal/logging"
)

// httpsDefaultPort is the IANA-standard HTTPS port. When the TLS listener
// is on this port the redirector omits the port from the redirect URL so
// browsers render the bare https://host form.
const httpsDefaultPort = 443

// httpToHTTPSRedirectHandler returns an [http.Handler] that 308-redirects
// every request to the equivalent HTTPS URL on httpsPort.
//
// 308 (Permanent Redirect, RFC 7538) is used rather than 301 so the HTTP
// method and body are preserved across the redirect. 301 allows clients
// to downgrade POST to GET, which would silently break state-changing
// API calls that arrive on the plaintext listener.
func httpToHTTPSRedirectHandler(httpsPort int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Build HTTPS URL preserving the host and path.
		host := r.Host
		// Strip any port from r.Host before re-joining with httpsPort —
		// otherwise the redirect target ends up like
		// https://localhost:8043:8444/... which browsers reject.
		if colonPos := strings.LastIndex(host, ":"); colonPos != -1 {
			// IPv6 literals look like "[::1]:8043"; only strip the port
			// if the colon is outside the bracketed host.
			if !strings.Contains(host[colonPos:], "]") {
				host = host[:colonPos]
			}
		}

		var httpsURL string
		if httpsPort == httpsDefaultPort {
			httpsURL = fmt.Sprintf("https://%s%s", host, r.RequestURI)
		} else {
			httpsURL = "https://" + net.JoinHostPort(host, strconv.Itoa(httpsPort)) + r.RequestURI
		}

		// #nosec G710 -- httpsURL is server-controlled: scheme/port from our config, host stripped to its
		// bare form before re-joining; user-supplied r.RequestURI is appended as the path/query only.
		http.Redirect(w, r, httpsURL, http.StatusPermanentRedirect)
	})
}

// startHTTPRedirect starts an HTTP server that 308-redirects all requests
// to HTTPS on httpsPort. Uses bindWithFallback so a busy redirectPort
// walks +1..+9 instead of killing the goroutine.
func (s *Server) startHTTPRedirect(redirectPort, httpsPort int) {
	ln, actualPort, bindErr := bindWithFallback(context.Background(), "", redirectPort)
	if bindErr != nil {
		logging.Error("HTTP→HTTPS redirect server bind failed",
			"requested_port", redirectPort, "error", bindErr)
		return
	}
	addr := fmt.Sprintf(":%d", actualPort)
	logging.Info("Starting HTTP→HTTPS redirect server",
		"addr", addr, "https_port", httpsPort)

	s.redirectServer = &http.Server{
		Addr:              addr,
		Handler:           httpToHTTPSRedirectHandler(httpsPort),
		ReadHeaderTimeout: redirectReadWriteTimeoutSec * time.Second,
		ReadTimeout:       redirectReadWriteTimeoutSec * time.Second,
		WriteTimeout:      redirectReadWriteTimeoutSec * time.Second,
	}

	err := s.redirectServer.Serve(ln)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Error("HTTP→HTTPS redirect server error", "error", err)
	}
}
