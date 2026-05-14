package handlers

import (
	"smart360/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCollectPrivateNotes(t *testing.T) {
	submissions := []models.Submission{
		{IsSelf: true, PrivateNotes: "should be ignored"},
		{Relationship: models.RelationshipManager, PrivateNotes: "thinks they're ready for L5"},
		{Relationship: models.RelationshipPeer, PrivateNotes: ""},
		{Relationship: models.RelationshipPeer, PrivateNotes: "   "},
		{Relationship: models.RelationshipPeer, PrivateNotes: "blocks on slack often"},
		{Relationship: models.RelationshipCrossFunctional, PrivateNotes: "needs PM hand-holding"},
	}

	notes := collectPrivateNotes(submissions)

	require.Len(t, notes, 3, "self + empty + whitespace notes are skipped")
	assert.Contains(t, notes[0], "manager")
	assert.Contains(t, notes[0], "thinks they're ready for L5")
	// peer and cross-functional both surface, with their relationship tag
	allJoined := notes[0] + "|" + notes[1] + "|" + notes[2]
	assert.Contains(t, allJoined, "peer")
	assert.Contains(t, allJoined, "cross-functional")
}

func TestBuildManagerOnlyChannel(t *testing.T) {
	t.Run("returns_nil_when_no_raw_notes", func(t *testing.T) {
		assert.Nil(t, buildManagerOnlyChannel(aiManagerOnlyPayload{Synthesis: "ignored"}, nil))
	})

	t.Run("attaches_count_and_raw_notes", func(t *testing.T) {
		ch := buildManagerOnlyChannel(
			aiManagerOnlyPayload{Synthesis: "Pattern observed.", Themes: []string{"theme one"}},
			[]string{"[peer] note 1", "[manager] note 2"},
		)
		require.NotNil(t, ch)
		assert.Equal(t, 2, ch.NoteCount)
		assert.Equal(t, "Pattern observed.", ch.Synthesis)
		assert.Equal(t, []string{"theme one"}, ch.Themes)
		assert.Equal(t, []string{"[peer] note 1", "[manager] note 2"}, ch.RawNotes)
	})
}

func TestCanSeeManagerOnlyChannel(t *testing.T) {
	creator := models.User{ID: primitive.NewObjectID(), Role: models.RoleMember}
	subject := models.User{ID: primitive.NewObjectID(), Role: models.RoleMember}
	admin := models.User{ID: primitive.NewObjectID(), Role: models.RoleAdmin}
	stranger := models.User{ID: primitive.NewObjectID(), Role: models.RoleTeamAdmin}
	round := models.FeedbackRound{CreatedByID: creator.ID, SubjectID: subject.ID}

	assert.True(t, canSeeManagerOnlyChannel(creator, round), "round creator can see private channel")
	assert.True(t, canSeeManagerOnlyChannel(admin, round), "global admin can see private channel")
	assert.False(t, canSeeManagerOnlyChannel(subject, round), "subject must never see private channel")
	assert.False(t, canSeeManagerOnlyChannel(stranger, round), "unrelated team admin can't see private channel")
}
