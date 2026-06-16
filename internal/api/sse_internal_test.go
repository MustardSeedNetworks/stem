package api

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api/sse"
)

// TestSSETransport_BroadcasterWiring verifies the transport-layer integration
// points: the Server struct holds an *sse.Broadcaster, and PublishTestProgress
// does not panic when called with no subscribers.
//
// Comprehensive broadcaster behaviour tests live in [internal/api/sse].
func TestSSETransport_BroadcasterWiring(_ *testing.T) {
	// Construct a minimal Server with only the broadcaster initialised so
	// we can exercise the wiring without needing the full NewServer auth
	// environment variables.
	s := &Server{sseBroadcaster: sse.New()}

	// Publishing with no subscribers must not panic.
	s.sseBroadcaster.Publish(sse.Frame{Type: "smoke", Payload: nil})
	s.PublishTestProgress("tid-1", map[string]string{"step": "smoke"})
}
