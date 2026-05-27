package api

import (
	"encoding/json"
	"sync"
	"time"
)

// SSE (server-sent events) broadcaster. Publishers call Publish() with a
// typed frame; every connected subscriber gets a copy via its channel.
//
// Slow subscribers are dropped (their channel is closed) rather than back-
// pressuring publishers — this keeps the broadcaster non-blocking from the
// publisher side, which matters because the mode-switch path runs under
// the request handler's goroutine and must not stall.
//
// Frame types are listed in the comment on Publish.

// SSEFrame is the wire form of an event frame. The `type` field is the
// discriminator the UI consumes; `payload` is the per-type body.
//
// Reflector stats and test progress frames replicate the structure of
// the matching REST responses so the consumer can drop the same
// rendering code into either source.
type SSEFrame struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// sseSubscriber owns one subscriber's outbound channel. Buffered so a
// brief client stall doesn't drop frames; capacity is bounded so a
// permanently-stalled client gets evicted instead of leaking memory.
type sseSubscriber struct {
	id uint64
	ch chan SSEFrame
}

const (
	// sseSubscriberBufferSize bounds the per-subscriber buffer; small
	// enough that a slow client drops within a couple of seconds at the
	// expected ~1Hz reflector-stats cadence, large enough that a brief
	// network blip doesn't trip eviction.
	sseSubscriberBufferSize = 16

	// sseHeartbeatInterval is how often the handler sends an SSE comment
	// line (": heartbeat\n\n") to keep idle proxies from closing the
	// connection. 15s is a common threshold.
	sseHeartbeatInterval = 15 * time.Second
)

// SSEBroadcaster fans out SSE frames to all connected subscribers.
//
// Process-wide singleton initialised once at server construction. The
// zero value is safe; New is for clarity at the call site.
type SSEBroadcaster struct {
	mu     sync.RWMutex
	subs   map[uint64]*sseSubscriber
	nextID uint64
}

// NewSSEBroadcaster returns a ready broadcaster with no subscribers.
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{subs: make(map[uint64]*sseSubscriber)}
}

// Subscribe registers a new subscriber and returns the channel to read
// frames from along with an unsubscribe function. The unsubscribe must
// be called (defer is the usual pattern) so the broadcaster doesn't
// hold a stale entry forever.
func (b *SSEBroadcaster) Subscribe() (<-chan SSEFrame, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	sub := &sseSubscriber{
		id: id,
		ch: make(chan SSEFrame, sseSubscriberBufferSize),
	}
	b.subs[id] = sub
	return sub.ch, func() { b.unsubscribe(id) }
}

func (b *SSEBroadcaster) unsubscribe(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	sub, ok := b.subs[id]
	if !ok {
		return
	}
	delete(b.subs, id)
	close(sub.ch)
}

// Publish fans out a frame to every subscriber. Slow subscribers (whose
// buffer is full) are dropped — their channel is closed and they're
// removed. Returns the number of subscribers that received the frame.
//
// Known frame types:
//
//   - "mode_changed": payload is ModeUpdateResponse (matches the
//     /api/v1/mode POST response)
//   - "reflector_stats": payload is ReflectorStats (matches
//     /api/v1/reflector/stats)
//   - "test_progress": payload is a per-test progress struct (TODO:
//     once test runner exposes a progress channel)
func (b *SSEBroadcaster) Publish(frame SSEFrame) int {
	// Hold the read lock across the sends so unsubscribe (which takes
	// the write lock to close the channel) can't race a "send on closed
	// channel" panic with us. Sends here are non-blocking (select +
	// default), so holding the read lock has bounded duration.
	b.mu.RLock()
	var stalled []uint64
	delivered := 0
	for _, sub := range b.subs {
		select {
		case sub.ch <- frame:
			delivered++
		default:
			// Subscriber buffer is full — they're stalled. Drop them
			// after we release the read lock so the broadcaster stays
			// non-blocking from the publisher's perspective.
			stalled = append(stalled, sub.id)
		}
	}
	b.mu.RUnlock()

	for _, id := range stalled {
		b.unsubscribe(id)
	}
	return delivered
}

// sseFrameOverhead is the byte count of the SSE wire-format wrapper
// (`data: ` prefix + `\n\n` suffix) we add around the JSON payload.
// Constant so the pre-allocation in Encode doesn't trip mnd.
const sseFrameOverhead = len("data: ") + len("\n\n")

// Encode renders a frame in SSE wire format: a `data:` line followed by
// the JSON payload and a terminating blank line. Multi-line JSON is
// guarded by serializing single-line.
//
// Returns the bytes including trailing "\n\n".
func (frame SSEFrame) Encode() ([]byte, error) {
	data, err := json.Marshal(frame)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(data)+sseFrameOverhead)
	out = append(out, "data: "...)
	out = append(out, data...)
	out = append(out, '\n', '\n')
	return out, nil
}
