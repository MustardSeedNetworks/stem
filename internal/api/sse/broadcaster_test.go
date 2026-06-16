// SPDX-License-Identifier: BUSL-1.1

package sse_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api/sse"
)

func TestBroadcaster_PublishToSubscribers(t *testing.T) {
	t.Parallel()
	b := sse.New()
	ch1, unsub1 := b.Subscribe()
	defer unsub1()
	ch2, unsub2 := b.Subscribe()
	defer unsub2()

	want := sse.Frame{Type: "mode_changed", Payload: map[string]string{"mode": "reflector"}}
	const expectedSubs = 2
	delivered := b.Publish(want)
	if delivered != expectedSubs {
		t.Errorf("expected delivery to %d subscribers, got %d", expectedSubs, delivered)
	}

	for i, ch := range []<-chan sse.Frame{ch1, ch2} {
		got, ok := <-ch
		if !ok {
			t.Errorf("subscriber %d: channel closed unexpectedly", i)
			continue
		}
		if got.Type != "mode_changed" {
			t.Errorf("subscriber %d: type = %q, want mode_changed", i, got.Type)
		}
	}
}

func TestBroadcaster_UnsubscribeStopsDelivery(t *testing.T) {
	t.Parallel()
	b := sse.New()
	ch, unsub := b.Subscribe()

	unsub()
	// After unsubscribe, the channel is closed.
	_, open := <-ch
	if open {
		t.Error("expected channel closed after unsubscribe")
	}

	// Publishing after unsubscribe should reach zero subscribers.
	delivered := b.Publish(sse.Frame{Type: "test_progress"})
	if delivered != 0 {
		t.Errorf("expected 0 deliveries after unsubscribe, got %d", delivered)
	}
}

func TestBroadcaster_SlowSubscriberDropped(t *testing.T) {
	t.Parallel()

	// subscriberBufferSize is not exported; use enough frames to definitely
	// overflow a buffer of 16 (the configured default).
	const overflowFrames = 20

	b := sse.New()
	ch, unsub := b.Subscribe()
	defer unsub()

	// Fill the subscriber's buffer beyond capacity so the next publish
	// finds it stalled and drops it.
	for i := range overflowFrames {
		b.Publish(sse.Frame{Type: "test", Payload: i})
	}

	// Drain. Channel should close eventually because subscriber was evicted.
	const maxAttempts = 100
	for attempt := range maxAttempts {
		select {
		case _, open := <-ch:
			if !open {
				// Channel closed — eviction happened, as expected.
				return
			}
		default:
			if attempt == maxAttempts-1 {
				t.Fatal("expected channel close after slow-subscriber eviction")
			}
		}
	}
}

func TestBroadcaster_ConcurrentSubscribePublish(_ *testing.T) {
	// Race-detector smoke test: many goroutines subscribing and publishing in
	// parallel must not race or panic. Subscribers drain on a best-effort basis
	// — we don't assert delivery counts, only that the operations complete
	// cleanly under -race.
	b := sse.New()
	var wg sync.WaitGroup

	const (
		publishers    = 30
		subscribers   = 10
		maxDrainPerCh = 3
	)
	for i := range publishers {
		wg.Go(func() {
			b.Publish(sse.Frame{Type: "burst", Payload: i})
		})
	}

	for range subscribers {
		wg.Go(func() {
			ch, unsub := b.Subscribe()
			drained := 0
		drain:
			for drained < maxDrainPerCh {
				select {
				case _, ok := <-ch:
					if !ok {
						break drain
					}
					drained++
				default:
					break drain
				}
			}
			unsub()
		})
	}

	wg.Wait()
}

func TestFrame_Encode(t *testing.T) {
	t.Parallel()
	frame := sse.Frame{Type: "mode_changed", Payload: map[string]string{"mode": "reflector"}}
	encoded, err := frame.Encode()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	s := string(encoded)
	if !strings.HasPrefix(s, "data: ") {
		t.Errorf("expected `data: ` prefix, got %q", s[:20])
	}
	if !strings.HasSuffix(s, "\n\n") {
		t.Errorf("expected trailing blank line (\\n\\n), got tail %q", s[len(s)-4:])
	}

	// The body between "data: " and "\n\n" is valid JSON.
	body := strings.TrimSuffix(strings.TrimPrefix(s, "data: "), "\n\n")
	var got sse.Frame
	if jsonErr := json.Unmarshal([]byte(body), &got); jsonErr != nil {
		t.Errorf("encoded body is not valid JSON: %v", jsonErr)
	}
	if got.Type != "mode_changed" {
		t.Errorf("decoded type = %q, want mode_changed", got.Type)
	}
}
