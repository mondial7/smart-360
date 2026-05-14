package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smart360/database"
	"smart360/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DownloadConsolidationPDF(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	roundID := c.Param("roundId")
	db := database.GetDB()
	ctx := context.Background()

	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	var round models.FeedbackRound
	if err := db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Authorization: subject of the round, the round creator, or any admin can download
	if currentUser.Role != models.RoleAdmin &&
		currentUser.ID != round.SubjectID &&
		currentUser.ID != round.CreatedByID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to access this consolidation"})
		return
	}

	var consolidation models.Consolidation
	if err := db.Collection("consolidations").FindOne(ctx, bson.M{"round_id": roundObjID}).Decode(&consolidation); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	// Subjects can only access shared consolidations
	if currentUser.Role != models.RoleAdmin && currentUser.ID == round.SubjectID && consolidation.SharedAt == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Consolidation has not been shared yet"})
		return
	}

	// Private manager-only channel must never reach the subject's PDF.
	if !canSeeManagerOnlyChannel(currentUser, round) {
		consolidation.ManagerOnlyChannel = nil
	}

	var subject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)

	pdfBytes, err := renderConsolidationPDF(subject, round, consolidation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render PDF"})
		return
	}

	filename := buildPDFFilename(subject.Name, round.CreatedAt.Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildPDFFilename(subjectName, dateStr string) string {
	slug := strings.ToLower(strings.ReplaceAll(subjectName, " ", "-"))
	if slug == "" {
		slug = "feedback"
	}
	return fmt.Sprintf("smart360-%s-%s.pdf", slug, dateStr)
}

func renderConsolidationPDF(subject models.User, round models.FeedbackRound, c models.Consolidation) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Header
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

	writeSection(pdf, "Key Strengths", parseJSONList(c.Strengths), true)
	writeSection(pdf, "Areas for Improvement", parseJSONList(c.AreasForImprovement), true)
	writeSection(pdf, "Actionable Insights", parseJSONList(c.ActionableInsights), true)

	if summaries := parseJSONStringMap(c.QuestionSummaries); len(summaries) > 0 {
		// Prefer the labels snapshotted from the template at generation time;
		// fall back to the conceptual defaults so legacy consolidations still
		// render readable headings.
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
			text, ok := summaries[key]
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
		pdf.MultiCell(0, 6, "⚠ Private — do NOT share with the subject.", "", "L", false)
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

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(160, 160, 160)
	pdf.CellFormat(0, 6, "Generated by Smart 360 Feedback - Confidential", "", 0, "C", false, 0, "")

	var buf strings.Builder
	if err := pdf.Output(&pdfWriter{sb: &buf}); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

type pdfWriter struct {
	sb *strings.Builder
}

func (w *pdfWriter) Write(p []byte) (int, error) {
	return w.sb.Write(p)
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

// competencyLines renders one row of the rubric PDF for each voice that
// actually submitted a rating. Skipping nil pointers keeps the PDF tight when
// a voice didn't contribute.
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

func parseJSONList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var list []string
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	return list
}

func parseJSONStringMap(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	return m
}

func fallback(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}
