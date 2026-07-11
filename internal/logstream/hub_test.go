package logstream

import "testing"

func TestHubRingBufferCap(t *testing.T) {
	h := New(3)
	for _, s := range []string{"a", "b", "c", "d"} {
		_, _ = h.Write([]byte(s + "\n"))
	}
	got := h.Recent()
	if len(got) != 3 || got[0] != "b" || got[2] != "d" {
		t.Fatalf("expected last 3 [b c d], got %v", got)
	}
}

func TestHubBroadcastsToSubscriber(t *testing.T) {
	h := New(10)
	sub, cancel := h.Subscribe()
	defer cancel()

	if _, err := h.Write([]byte("line one\nline two\n")); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"line one", "line two"} {
		select {
		case got := <-sub:
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		default:
			t.Fatalf("expected %q on the subscriber channel", want)
		}
	}
}

func TestHubWriteSplitsAndSkipsBlank(t *testing.T) {
	h := New(10)
	_, _ = h.Write([]byte("x\n\n  \ny\n"))
	got := h.Recent()
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("expected [x y], got %v", got)
	}
}
