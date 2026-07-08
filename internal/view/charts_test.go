package view

import (
	"math"
	"strings"
	"testing"
)

func TestRadarSVG(t *testing.T) {
	svg := string(RadarSVG([]RadarAxis{
		{Label: "Execution", Value: 4},
		{Label: "Collaboration", Value: 3},
		{Label: "Ownership", Value: 5},
	}, 240, 5))

	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatalf("expected a complete svg element, got: %.40s...", svg)
	}
	// One data polygon (radar__shape) and one point per axis.
	if strings.Count(svg, "radar__shape") != 1 {
		t.Fatalf("expected exactly one data polygon")
	}
	if got := strings.Count(svg, `class="radar__point"`); got != 3 {
		t.Fatalf("expected 3 data points, got %d", got)
	}
	// Labels are escaped and carry the value.
	if !strings.Contains(svg, "Execution (4)") {
		t.Fatalf("expected labelled axis, got: %s", svg)
	}
}

func TestRadarSVG_Empty(t *testing.T) {
	if RadarSVG(nil, 240, 0) != "" {
		t.Fatal("expected empty output for no axes")
	}
}

func TestDonutSVG(t *testing.T) {
	svg := string(DonutSVG([]DonutSlice{
		{Label: "Active", Value: 3, Color: "#4f46e5"},
		{Label: "Closed", Value: 1, Color: "#16a34a"},
		{Label: "Draft", Value: 0, Color: "#999"},
	}, 180, 24, "rounds"))

	// Two non-zero slices → two segment circles (zero-valued slice skipped).
	if got := strings.Count(svg, "donut__segment"); got != 2 {
		t.Fatalf("expected 2 segments, got %d", got)
	}
	// Center total is the sum.
	if !strings.Contains(svg, ">4</text>") {
		t.Fatalf("expected total of 4 in center, got: %s", svg)
	}
	if !strings.Contains(svg, ">rounds</text>") {
		t.Fatalf("expected center label, got: %s", svg)
	}
	// Slice color is present.
	if !strings.Contains(svg, "#4f46e5") {
		t.Fatalf("expected slice color in output")
	}
}

func TestDonutSVG_AllZero(t *testing.T) {
	svg := string(DonutSVG([]DonutSlice{{Label: "None", Value: 0, Color: "#000"}}, 180, 24, "rounds"))
	if strings.Contains(svg, "donut__segment") {
		t.Fatal("expected no segments when all slices are zero")
	}
	if !strings.Contains(svg, ">0</text>") {
		t.Fatal("expected zero total")
	}
}

func TestNum(t *testing.T) {
	cases := map[float64]string{120: "120", 63.636: "63.64", 0: "0", 12.5: "12.5"}
	for in, want := range cases {
		if got := num(in); got != want {
			t.Errorf("num(%v) = %q, want %q", in, got, want)
		}
	}
	// Negative zero must normalize to "0", not "-0".
	if got := num(math.Copysign(0, -1)); got != "0" {
		t.Errorf("num(-0) = %q, want %q", got, "0")
	}
}
