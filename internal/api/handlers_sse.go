package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// handleSSEEvents serves the long-lived SSE stream at /api/v1/events.
//
// Wire protocol: each frame is one JSON object preceded by `data: ` and
// followed by a blank line. Periodic SSE-comment heartbeats keep
// intermediaries (load balancers, reverse proxies) from idling the
// connection — they discard comments so the UI never sees them.
//
// The handler blocks for the lifetime of the connection. It exits when
// either the client disconnects (r.Context().Done()) or the connection
// breaks on a write. No reconnect logic on the server side; the
// EventSource API on the browser handles reconnection automatically.
func (s *Server) handleSSEEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	// SSE requires a flusher so we can stream frames as they arrive.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Disable the server-level WriteTimeout for this connection — SSE
	// is long-lived by design. http.ResponseController (Go 1.20+) lets
	// us set a zero deadline (= no timeout) on a per-request basis
	// without affecting other endpoints' timeouts.
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		// Not fatal — SSE will still work until the global WriteTimeout
		// fires, but the connection will then break and the client will
		// auto-reconnect. Log so operators see the warning.
		logging.FromContext(r.Context()).WarnContext(r.Context(),
			"SSE: failed to clear write deadline", "error", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Tell intermediary proxies (nginx default) not to buffer.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Initial connection-established comment lets the client know it's
	// live without waiting for the first real frame. EventSource fires
	// `open` on first byte received.
	if _, err := fmt.Fprintf(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	frames, unsubscribe := s.sseBroadcaster.Subscribe()
	defer unsubscribe()

	heartbeat := time.NewTicker(sseHeartbeatInterval)
	defer heartbeat.Stop()

	ctx := r.Context()
	logger := logging.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			// Client disconnected — exit cleanly.
			return

		case frame, open := <-frames:
			if !open {
				// Broadcaster dropped us (buffer overflow). Tell the
				// client we're closing so it reconnects.
				_, _ = fmt.Fprintf(w, ": evicted\n\n")
				flusher.Flush()
				return
			}
			payload, err := frame.Encode()
			if err != nil {
				logger.WarnContext(ctx, "SSE encode failed",
					"type", frame.Type, "error", err)
				continue
			}
			if _, writeErr := w.Write(payload); writeErr != nil {
				// Connection broken — exit. The client's EventSource
				// will reconnect on its own schedule.
				return
			}
			flusher.Flush()

		case <-heartbeat.C:
			// SSE comment lines start with `:` and are ignored by the
			// client. Used purely to keep proxies happy.
			if _, err := fmt.Fprintf(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
