package ai

import (
	"encoding/json"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

type aiDeltaPayload struct {
	SelfSubmitted   bool     `json:"self_submitted"`
	BlindSpots      []string `json:"blind_spots"`
	HiddenStrengths []string `json:"hidden_strengths"`
	Aligned         []string `json:"aligned"`
	Summary         string   `json:"summary"`
}

type aiVoicePayload struct {
	Summary string   `json:"summary"`
	Themes  []string `json:"themes"`
}

type aiVoiceBreakdownPayload struct {
	ManagerVoice aiVoicePayload `json:"manager_voice"`
	PeerVoice    aiVoicePayload `json:"peer_voice"`
	ReportVoice  aiVoicePayload `json:"report_voice"`
}

type aiManagerOnlyPayload struct {
	Synthesis string   `json:"synthesis"`
	Themes    []string `json:"themes"`
}

// aiPayload is the full JSON shape expected from Gemini.
type aiPayload struct {
	ExecutiveSummary    string                  `json:"executive_summary"`
	Strengths           []string                `json:"strengths"`
	AreasForImprovement []string                `json:"areas_for_improvement"`
	ActionableInsights  []string                `json:"actionable_insights"`
	QuestionSummaries   map[string]string       `json:"question_summaries"`
	SelfVsOthersDelta   aiDeltaPayload          `json:"self_vs_others_delta"`
	VoiceBreakdown      aiVoiceBreakdownPayload `json:"voice_breakdown"`
	ManagerOnlyChannel  aiManagerOnlyPayload    `json:"manager_only_channel"`
}

// parseAIPayload extracts and parses the JSON object from a model response,
// tolerating code fences or surrounding prose. On any failure it returns a
// safe fallback payload so the consolidation still persists something useful.
func parseAIPayload(responseText string, hasSelf bool) aiPayload {
	clean := strings.TrimSpace(responseText)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```JSON")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	if i := strings.IndexRune(clean, '{'); i != -1 {
		if j := strings.LastIndex(clean, "}"); j != -1 && j > i {
			clean = clean[i : j+1]
		}
	}

	var p aiPayload
	if !json.Valid([]byte(clean)) {
		return fallbackAIPayload(hasSelf)
	}
	if err := json.Unmarshal([]byte(clean), &p); err != nil {
		return fallbackAIPayload(hasSelf)
	}
	return p
}

func fallbackAIPayload(hasSelf bool) aiPayload {
	p := aiPayload{
		ExecutiveSummary:    "Consolidation could not be generated automatically. Please review the raw submissions.",
		Strengths:           []string{},
		AreasForImprovement: []string{},
		ActionableInsights:  []string{},
		QuestionSummaries:   map[string]string{"a": "", "b": "", "c": "", "d": ""},
	}
	p.SelfVsOthersDelta.SelfSubmitted = hasSelf
	p.SelfVsOthersDelta.BlindSpots = []string{}
	p.SelfVsOthersDelta.HiddenStrengths = []string{}
	p.SelfVsOthersDelta.Aligned = []string{}
	p.VoiceBreakdown.ManagerVoice.Themes = []string{}
	p.VoiceBreakdown.PeerVoice.Themes = []string{}
	p.VoiceBreakdown.ReportVoice.Themes = []string{}
	p.ManagerOnlyChannel.Themes = []string{}
	return p
}

func voiceSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"summary": {Type: genai.TypeString},
			"themes":  {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
		},
		Required: []string{"summary", "themes"},
	}
}

func consolidationSchema() *genai.Schema {
	strArray := &genai.Schema{Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}}
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"executive_summary":     {Type: genai.TypeString},
			"strengths":             strArray,
			"areas_for_improvement": strArray,
			"actionable_insights":   strArray,
			"question_summaries": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"a": {Type: genai.TypeString},
					"b": {Type: genai.TypeString},
					"c": {Type: genai.TypeString},
					"d": {Type: genai.TypeString},
				},
				Required: []string{"a", "b", "c", "d"},
			},
			"self_vs_others_delta": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"self_submitted":   {Type: genai.TypeBoolean},
					"blind_spots":      strArray,
					"hidden_strengths": strArray,
					"aligned":          strArray,
					"summary":          {Type: genai.TypeString},
				},
				Required: []string{"self_submitted", "blind_spots", "hidden_strengths", "aligned", "summary"},
			},
			"voice_breakdown": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"manager_voice": voiceSchema(),
					"peer_voice":    voiceSchema(),
					"report_voice":  voiceSchema(),
				},
				Required: []string{"manager_voice", "peer_voice", "report_voice"},
			},
			"manager_only_channel": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"synthesis": {Type: genai.TypeString},
					"themes":    strArray,
				},
				Required: []string{"synthesis", "themes"},
			},
		},
		Required: []string{"executive_summary", "strengths", "areas_for_improvement", "actionable_insights", "question_summaries", "self_vs_others_delta", "voice_breakdown", "manager_only_channel"},
	}
}
