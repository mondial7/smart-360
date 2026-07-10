// Package pdf renders a consolidation to a printable PDF using fpdf. It is
// intentionally dependency-light: no database, no HTTP. Visibility rules (e.g.
// stripping the manager-only channel for the subject) are applied by the caller
// before handing the consolidation here.
package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/mondial7/smart-360/internal/models"
)

// Filename builds the download filename, e.g. "smart360-jane-doe-2026-07-08.pdf".
func Filename(subjectName, dateStr string) string {
	slug := strings.ToLower(strings.ReplaceAll(subjectName, " ", "-"))
	if slug == "" {
		slug = "feedback"
	}
	return fmt.Sprintf("smart360-%s-%s.pdf", slug, dateStr)
}

// Render produces the consolidation PDF bytes for a subject and round.
func Render(subject models.User, round models.FeedbackRound, c models.Consolidation) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetTextColor(20, 30, 60)
	pdf.CellFormat(0, 12, "Smart 360 Feedback", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, fmt.Sprintf("Subject: %s", fallback(subject.Name, "Unknown")), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Round date: %s", round.CreatedAt.Format("January 2, 2006")), "", 1, "L", false, 0, "")
	if c.SharedAt != nil {
		pdf.CellFormat(0, 6, fmt.Sprintf("Shared on: %s", c.SharedAt.Format("January 2, 2006")), "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	if c.ExecutiveSummary != "" {
		writeSection(pdf, "Executive Summary", []string{c.ExecutiveSummary}, false)
	}

	writeSection(pdf, "Key Strengths", c.Strengths, true)
	writeSection(pdf, "Areas for Improvement", c.AreasForImprovement, true)
	writeSection(pdf, "Actionable Insights", c.ActionableInsights, true)

	if len(c.QuestionSummaries) > 0 {
		labels := map[string]string{
			"a": "What to continue",
			"b": "What's blocking growth",
			"c": "Where to double down",
			"d": "One experiment to try (30–60 days)",
		}
		for k, v := range c.QuestionLabels {
			if strings.TrimSpace(v) != "" {
				labels[k] = v
			}
		}
		writeSectionTitle(pdf, "Question-by-Question Summary")
		for _, key := range []string{"a", "b", "c", "d"} {
			text, ok := c.QuestionSummaries[key]
			if !ok || strings.TrimSpace(text) == "" {
				continue
			}
			pdf.SetFont("Helvetica", "B", 11)
			pdf.SetTextColor(40, 40, 40)
			pdf.CellFormat(0, 6, labels[key], "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 11)
			pdf.SetTextColor(60, 60, 60)
			pdf.MultiCell(0, 6, text, "", "L", false)
			pdf.Ln(2)
		}
	}

	if len(c.CompetencyRatings) > 0 {
		writeSectionTitle(pdf, "Competency Ratings")
		for _, r := range c.CompetencyRatings {
			pdf.SetFont("Helvetica", "B", 11)
			pdf.SetTextColor(40, 40, 40)
			heading := r.Name
			if r.OthersAverage != nil {
				heading = fmt.Sprintf("%s — %.1f / 5 avg (%d reviewers)", r.Name, *r.OthersAverage, r.OthersCount)
			}
			pdf.CellFormat(0, 6, heading, "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(60, 60, 60)
			for _, line := range competencyLines(r) {
				pdf.MultiCell(0, 5, "  "+line, "", "L", false)
			}
			if r.Spread >= 2 {
				pdf.SetFont("Helvetica", "I", 10)
				pdf.SetTextColor(150, 75, 0)
				pdf.MultiCell(0, 5, fmt.Sprintf("  Wide spread (%.1f points) — reviewers disagree on this axis.", r.Spread), "", "L", false)
				pdf.SetTextColor(60, 60, 60)
			}
			pdf.Ln(2)
		}
	}

	if c.VoiceBreakdown != nil {
		writeSectionTitle(pdf, "Voices — Different Vantages, Distinct Signals")
		writeVoiceBlock(pdf, "Manager voice", c.VoiceBreakdown.ManagerVoice)
		writeVoiceBlock(pdf, "Peer voice", c.VoiceBreakdown.PeerVoice)
		writeVoiceBlock(pdf, "Direct-report voice", c.VoiceBreakdown.ReportVoice)
	}

	if c.SelfVsOthersDelta != nil && c.SelfVsOthersDelta.SelfSubmitted {
		writeSectionTitle(pdf, "Self vs Peers — Where You and Your Team See Things Differently")
		if c.SelfVsOthersDelta.Summary != "" {
			pdf.SetFont("Helvetica", "I", 11)
			pdf.SetTextColor(60, 60, 60)
			pdf.MultiCell(0, 6, c.SelfVsOthersDelta.Summary, "", "L", false)
			pdf.Ln(3)
		}
		writeDeltaSubsection(pdf, "Blind spots (peers see, you may not)", c.SelfVsOthersDelta.BlindSpots)
		writeDeltaSubsection(pdf, "Hidden strengths (you may underestimate)", c.SelfVsOthersDelta.HiddenStrengths)
		writeDeltaSubsection(pdf, "Aligned themes (self and peers agree)", c.SelfVsOthersDelta.Aligned)
	}

	if c.ManagerOnlyChannel != nil {
		writeSectionTitle(pdf, fmt.Sprintf("Manager-Only Channel (%d private note(s))", c.ManagerOnlyChannel.NoteCount))
		pdf.SetFont("Helvetica", "I", 11)
		pdf.SetTextColor(150, 75, 0)
		pdf.MultiCell(0, 6, "Private — do NOT share with the subject.", "", "L", false)
		pdf.Ln(2)
		if c.ManagerOnlyChannel.Synthesis != "" {
			pdf.SetFont("Helvetica", "", 11)
			pdf.SetTextColor(60, 60, 60)
			pdf.MultiCell(0, 6, c.ManagerOnlyChannel.Synthesis, "", "L", false)
			pdf.Ln(2)
		}
		writeDeltaSubsection(pdf, "Themes", c.ManagerOnlyChannel.Themes)
		writeDeltaSubsection(pdf, "Raw notes (relationship-tagged, anonymous)", c.ManagerOnlyChannel.RawNotes)
	}

	if c.AdminNotes != "" {
		writeSection(pdf, "Admin Notes", []string{c.AdminNotes}, false)
	}

	pdf.SetY(-20)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(160, 160, 160)
	pdf.CellFormat(0, 6, "Generated by Smart 360 Feedback - Confidential", "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(20, 30, 60)
	pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	pdf.Ln(1)
}

func writeSection(pdf *fpdf.Fpdf, title string, items []string, bullet bool) {
	if len(items) == 0 {
		return
	}
	writeSectionTitle(pdf, title)
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(60, 60, 60)
	for _, item := range items {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		if bullet {
			text = "- " + text
		}
		pdf.MultiCell(0, 6, text, "", "L", false)
		pdf.Ln(1)
	}
	pdf.Ln(3)
}

func competencyLines(r models.CompetencyRatingAggregate) []string {
	var out []string
	if r.SelfScore != nil {
		out = append(out, fmt.Sprintf("Self:    %.1f", *r.SelfScore))
	}
	if r.ManagerAverage != nil {
		out = append(out, fmt.Sprintf("Manager: %.1f", *r.ManagerAverage))
	}
	if r.PeerAverage != nil {
		out = append(out, fmt.Sprintf("Peers:   %.1f", *r.PeerAverage))
	}
	if r.ReportAverage != nil {
		out = append(out, fmt.Sprintf("Reports: %.1f", *r.ReportAverage))
	}
	return out
}

func writeVoiceBlock(pdf *fpdf.Fpdf, title string, v *models.Voice) {
	if v == nil {
		return
	}
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(0, 6, fmt.Sprintf("%s — %d reviewer(s)", title, v.ReviewerCount), "", 1, "L", false, 0, "")
	if v.Summary != "" {
		pdf.SetFont("Helvetica", "I", 11)
		pdf.SetTextColor(60, 60, 60)
		pdf.MultiCell(0, 6, v.Summary, "", "L", false)
	}
	if len(v.Themes) > 0 {
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(60, 60, 60)
		for _, theme := range v.Themes {
			t := strings.TrimSpace(theme)
			if t == "" {
				continue
			}
			pdf.MultiCell(0, 6, "- "+t, "", "L", false)
		}
	}
	pdf.Ln(2)
}

func writeDeltaSubsection(pdf *fpdf.Fpdf, title string, items []string) {
	if len(items) == 0 {
		return
	}
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(0, 6, title, "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(60, 60, 60)
	for _, item := range items {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		pdf.MultiCell(0, 6, "- "+text, "", "L", false)
	}
	pdf.Ln(2)
}

func fallback(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}
