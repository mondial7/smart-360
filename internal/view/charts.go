// Package view holds server-side rendering helpers: the template registry and
// the SVG chart generators ported from the former Vue components. Charts are
// emitted as plain SVG markup so no client-side chart library is needed.
package view

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"
)

// RadarAxis is one spoke of the radar chart.
type RadarAxis struct {
	Label string
	Value float64
}

// RadarSVG renders a radar/spider chart as SVG. size is the square viewport in
// px; max pins the outer ring (0 → auto-scale to 110% of the largest value).
// Geometry mirrors the former RadarChart.vue exactly.
func RadarSVG(axes []RadarAxis, size, max float64) template.HTML {
	if len(axes) == 0 {
		return ""
	}
	if size <= 0 {
		size = 240
	}
	center := size / 2
	radius := size/2 - 28
	count := float64(len(axes))

	effectiveMax := max
	if effectiveMax <= 0 {
		dataMax := 1.0
		for _, a := range axes {
			if a.Value > dataMax {
				dataMax = a.Value
			}
		}
		effectiveMax = math.Ceil(dataMax * 1.1)
	}
	if effectiveMax == 0 {
		effectiveMax = 1
	}

	point := func(i int, value float64) (float64, float64) {
		angle := (2*math.Pi*float64(i))/count - math.Pi/2
		r := radius * (value / effectiveMax)
		return center + r*math.Cos(angle), center + r*math.Sin(angle)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %s %s" width="%s" height="%s" role="img" aria-label="Feedback radar chart" class="radar">`,
		num(size), num(size), num(size), num(size))

	// Rings.
	b.WriteString(`<g class="radar__rings">`)
	for _, ratio := range []float64{0.25, 0.5, 0.75, 1} {
		var pts []string
		for i := range axes {
			angle := (2*math.Pi*float64(i))/count - math.Pi/2
			r := radius * ratio
			pts = append(pts, num(center+r*math.Cos(angle))+","+num(center+r*math.Sin(angle)))
		}
		fmt.Fprintf(&b, `<polygon points="%s" class="radar__ring" />`, strings.Join(pts, " "))
	}
	b.WriteString(`</g>`)

	// Axis spokes.
	b.WriteString(`<g class="radar__axes">`)
	for i := range axes {
		x, y := point(i, effectiveMax)
		fmt.Fprintf(&b, `<line x1="%s" y1="%s" x2="%s" y2="%s" class="radar__axis" />`,
			num(center), num(center), num(x), num(y))
	}
	b.WriteString(`</g>`)

	// Data polygon.
	var poly []string
	for i, a := range axes {
		x, y := point(i, a.Value)
		poly = append(poly, num(x)+","+num(y))
	}
	fmt.Fprintf(&b, `<polygon points="%s" class="radar__shape" />`, strings.Join(poly, " "))

	// Data points.
	b.WriteString(`<g class="radar__points">`)
	for i, a := range axes {
		x, y := point(i, a.Value)
		fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="3.5" class="radar__point" />`, num(x), num(y))
	}
	b.WriteString(`</g>`)

	// Labels.
	b.WriteString(`<g class="radar__labels">`)
	for i, a := range axes {
		angle := (2*math.Pi*float64(i))/count - math.Pi/2
		r := radius + 18
		lx := center + r*math.Cos(angle)
		ly := center + r*math.Sin(angle)
		fmt.Fprintf(&b, `<text x="%s" y="%s" text-anchor="middle" dominant-baseline="middle" class="radar__label">%s (%s)</text>`,
			num(lx), num(ly), template.HTMLEscapeString(a.Label), num(a.Value))
	}
	b.WriteString(`</g></svg>`)

	return template.HTML(b.String()) // #nosec G203 -- markup built from escaped labels + numeric geometry
}

// DonutSlice is one wedge of the donut chart.
type DonutSlice struct {
	Label string
	Value int
	Color string
}

// DonutSVG renders a donut chart as SVG with a centered total and label.
// Geometry mirrors the former StatusDonut.vue (segments as stroke-dasharray on
// concentric circles, rotated -90° to start at 12 o'clock).
func DonutSVG(slices []DonutSlice, size, thickness float64, centerLabel string) template.HTML {
	if size <= 0 {
		size = 180
	}
	if thickness <= 0 {
		thickness = 24
	}
	total := 0
	for _, s := range slices {
		total += s.Value
	}
	radius := (size - thickness) / 2
	circumference := 2 * math.Pi * radius
	center := size / 2

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %s %s" width="%s" height="%s" class="donut__svg" role="img" aria-label="Status breakdown">`,
		num(size), num(size), num(size), num(size))

	// Track ring.
	fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="%s" fill="none" stroke="var(--bg-secondary)" stroke-width="%s" />`,
		num(center), num(center), num(radius), num(thickness))

	// Segments.
	if total > 0 {
		consumed := 0.0
		for _, s := range slices {
			if s.Value <= 0 {
				continue
			}
			length := (float64(s.Value) / float64(total)) * circumference
			fmt.Fprintf(&b,
				`<circle cx="%s" cy="%s" r="%s" fill="none" stroke="%s" stroke-width="%s" stroke-dasharray="%s %s" stroke-dashoffset="%s" transform="rotate(-90 %s %s)" class="donut__segment" />`,
				num(center), num(center), num(radius), template.HTMLEscapeString(s.Color), num(thickness),
				num(length), num(circumference-length), num(-consumed), num(center), num(center))
			consumed += length
		}
	}

	fmt.Fprintf(&b, `<text x="%s" y="%s" text-anchor="middle" dominant-baseline="middle" class="donut__total-value">%d</text>`,
		num(center), num(center-4), total)
	fmt.Fprintf(&b, `<text x="%s" y="%s" text-anchor="middle" dominant-baseline="middle" class="donut__total-label">%s</text>`,
		num(center), num(center+14), template.HTMLEscapeString(centerLabel))
	b.WriteString(`</svg>`)

	return template.HTML(b.String()) // #nosec G203 -- markup built from escaped labels + numeric geometry
}

// num formats a float for SVG coordinates: up to 2 decimals, trailing zeros
// trimmed, so "120" stays "120" and "63.64" stays compact.
func num(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}
