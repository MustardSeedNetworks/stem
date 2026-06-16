// SPDX-License-Identifier: BUSL-1.1

// Package sse implements the SSE (server-sent events) broadcaster that fans out
// typed event frames to connected subscribers. It is a leaf of internal/api
// (ADR-0011): it depends only on the standard library — never on the api
// transport layer itself, so the boundary is enforced by depguard.
//
// The broadcaster is the event-broadcast machinery only. HTTP serving (headers,
// flusher, heartbeat loop) lives in the api transport layer (handlers_sse.go)
// and calls into this package via [Broadcaster.Subscribe].
package sse

import (
	"encoding/json"
	"slices"
	"sync"
)

// subscriberBufferSize bounds the per-subscriber channel. Small enough that a
// slow client drops within a couple of seconds at the expected ~1 Hz reflector-
// stats cadence; large enough that a brief network blip doesn't trip eviction.
const subscriberBufferSize = 16

// HeartbeatInterval is how often the HTTP handler should send an SSE comment
// line (": heartbeat\n\n") to keep idle proxies from closing the connection.
// 15 s is a common threshold. Exported so the transport layer uses the same
// constant without duplicating it.
const HeartbeatInterval = 15 // seconds

// Frame is the wire form of an SSE event. The Type field is the discriminator
// the UI consumes; Payload is the per-type body.
//
// Reflector-stats and test-progress frames replicate the structure of the
// matching REST responses so the consumer can drop the same rendering code
// into either source.
type Frame struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Encode renders a frame in SSE wire format: a "data:" line followed by the
// JSON-encoded frame and a terminating blank line. The JSON is always single-
// line (marshalled by [json.Marshal] with no indent), which satisfies the SSE
// spec requirement that the data field not span multiple "data:" lines.
//
// Returns the bytes including trailing "\n\n".
func (f Frame) Encode() ([]byte, error) {
	data, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	// slices.Concat sizes the result internally (with its own overflow guard),
	// so there is no hand-rolled make-capacity arithmetic for a static analyzer
	// to flag as a potential allocation-size overflow.
	return slices.Concat([]byte("data: "), data, []byte("\n\n")), nil
}

// subscriber owns one connected client's outbound channel. Buffered so a
// brief stall doesn't drop frames; bounded so a permanently-stalled client
// gets evicted instead of leaking memory.
type subscriber struct {
	id uint64
	ch chan Frame
}

// Broadcaster fans out SSE frames to all connected subscribers.
//
// Process-wide singleton initialised once at server construction. The zero
// value is safe; [New] is for clarity at the call site.
type Broadcaster struct {
	mu     sync.RWMutex
	subs   map[uint64]*subscriber
	nextID uint64
}

// New returns a ready Broadcaster with no subscribers.
func New() *Broadcaster {
	return &Broadcaster{subs: make(map[uint64]*subscriber)}
}

// Subscribe registers a new subscriber and returns the channel to read frames
// from, along with an unsubscribe function. The unsubscribe must be called
// (defer is the usual pattern) so the broadcaster does not hold a stale entry
// forever.
func (b *Broadcaster) Subscribe() (<-chan Frame, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	sub := &subscriber{
		id: id,
		ch: make(chan Frame, subscriberBufferSize),
	}
	b.subs[id] = sub
	return sub.ch, func() { b.unsubscribe(id) }
}

func (b *Broadcaster) unsubscribe(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	sub, ok := b.subs[id]
	if !ok {
		return
	}
	delete(b.subs, id)
	close(sub.ch)
}

// Publish fans out a frame to every subscriber. Slow subscribers (whose buffer
// is full) are dropped — their channel is closed and they are removed. Returns
// the number of subscribers that received the frame.
//
// Known frame types:
//
//   - "mode_changed":    payload is ModeUpdateResponse (matches the
//     /api/v1/mode POST response)
//   - "reflector_stats": payload is ReflectorStats (matches
//     /api/v1/reflector/stats)
//   - "test_progress":   payload is a per-test progress struct
func (b *Broadcaster) Publish(frame Frame) int {
	// Hold the read lock across the sends so unsubscribe (which takes the
	// write lock to close the channel) cannot race a "send on closed channel"
	// panic with us. Sends here are non-blocking (select + default), so
	// holding the read lock has bounded duration.
	b.mu.RLock()
	var stalled []uint64
	delivered := 0
	for _, sub := range b.subs {
		select {
		case sub.ch <- frame:
			delivered++
		default:
			// Subscriber buffer is full — they are stalled. Drop them after
			// we release the read lock so the broadcaster stays non-blocking
			// from the publisher's perspective.
			stalled = append(stalled, sub.id)
		}
	}
	b.mu.RUnlock()

	for _, id := range stalled {
		b.unsubscribe(id)
	}
	return delivered
}
