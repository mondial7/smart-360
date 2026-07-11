// Package logstream captures application log lines into a bounded ring buffer
// and fans them out to live subscribers. It is an io.Writer, so it slots in
// behind slog and the standard logger without changing call sites, and it backs
// the admin "Logs" page (live stream in the browser) plus any consumer that
// connects to the log-stream endpoint.
package logstream

import (
	"strings"
	"sync"
)

// Hub buffers recent log lines and broadcasts new ones to subscribers.
type Hub struct {
	mu   sync.Mutex
	ring []string
	size int
	subs map[chan string]struct{}
}

// New returns a Hub retaining the last size lines (default 500).
func New(size int) *Hub {
	if size <= 0 {
		size = 500
	}
	return &Hub{size: size, subs: make(map[chan string]struct{})}
}

// Write implements io.Writer: each newline-delimited line is buffered and
// broadcast. It never blocks the caller (slow subscribers drop lines).
func (h *Hub) Write(p []byte) (int, error) {
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		h.publish(line)
	}
	return len(p), nil
}

func (h *Hub) publish(line string) {
	h.mu.Lock()
	h.ring = append(h.ring, line)
	if len(h.ring) > h.size {
		h.ring = h.ring[len(h.ring)-h.size:]
	}
	subs := make([]chan string, 0, len(h.subs))
	for c := range h.subs {
		subs = append(subs, c)
	}
	h.mu.Unlock()

	for _, c := range subs {
		select {
		case c <- line:
		default: // subscriber is behind; drop rather than block logging
		}
	}
}

// Recent returns a copy of the buffered lines, oldest first.
func (h *Hub) Recent() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, len(h.ring))
	copy(out, h.ring)
	return out
}

// Subscribe registers a live subscriber and returns its channel plus a cancel
// func. The channel is buffered; the cancel func unregisters it (it is not
// closed, so a concurrent publish can never send on a closed channel — readers
// stop via their own context).
func (h *Hub) Subscribe() (<-chan string, func()) {
	c := make(chan string, 128)
	h.mu.Lock()
	h.subs[c] = struct{}{}
	h.mu.Unlock()
	return c, func() {
		h.mu.Lock()
		delete(h.subs, c)
		h.mu.Unlock()
	}
}
