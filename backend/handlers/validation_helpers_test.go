package handlers

import (
	"smart360/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStatusTransition(t *testing.T) {
	tests := []struct {
		name      string
		current   models.RoundStatus
		new       models.RoundStatus
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid_draft_to_active",
			current:   models.RoundDraft,
			new:       models.RoundActive,
			expectErr: false,
		},
		{
			name:      "invalid_draft_to_closed",
			current:   models.RoundDraft,
			new:       models.RoundClosed,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "invalid_draft_to_shared",
			current:   models.RoundDraft,
			new:       models.RoundShared,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "valid_active_to_closed",
			current:   models.RoundActive,
			new:       models.RoundClosed,
			expectErr: false,
		},
		{
			name:      "invalid_active_to_draft",
			current:   models.RoundActive,
			new:       models.RoundDraft,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "invalid_active_to_shared",
			current:   models.RoundActive,
			new:       models.RoundShared,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "valid_closed_to_shared",
			current:   models.RoundClosed,
			new:       models.RoundShared,
			expectErr: false,
		},
		{
			name:      "invalid_closed_to_draft",
			current:   models.RoundClosed,
			new:       models.RoundDraft,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "invalid_closed_to_active",
			current:   models.RoundClosed,
			new:       models.RoundActive,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "invalid_shared_to_any_status",
			current:   models.RoundShared,
			new:       models.RoundDraft,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "invalid_shared_to_active",
			current:   models.RoundShared,
			new:       models.RoundActive,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
		{
			name:      "invalid_shared_to_closed",
			current:   models.RoundShared,
			new:       models.RoundClosed,
			expectErr: true,
			errCode:   "INVALID_STATUS_TRANSITION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStatusTransition(tt.current, tt.new)

			if tt.expectErr {
				require.Error(t, err)
				validationErr, ok := err.(*ValidationError)
				require.True(t, ok, "error should be a ValidationError")
				assert.Equal(t, tt.errCode, validationErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSubjectChange(t *testing.T) {
	tests := []struct {
		name      string
		status    models.RoundStatus
		expectErr bool
		errCode   string
	}{
		{
			name:      "valid_subject_change_in_draft",
			status:    models.RoundDraft,
			expectErr: false,
		},
		{
			name:      "invalid_subject_change_in_active",
			status:    models.RoundActive,
			expectErr: true,
			errCode:   "INVALID_SUBJECT_CHANGE",
		},
		{
			name:      "invalid_subject_change_in_closed",
			status:    models.RoundClosed,
			expectErr: true,
			errCode:   "INVALID_SUBJECT_CHANGE",
		},
		{
			name:      "invalid_subject_change_in_shared",
			status:    models.RoundShared,
			expectErr: true,
			errCode:   "INVALID_SUBJECT_CHANGE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSubjectChange(tt.status)

			if tt.expectErr {
				require.Error(t, err)
				validationErr, ok := err.(*ValidationError)
				require.True(t, ok, "error should be a ValidationError")
				assert.Equal(t, tt.errCode, validationErr.Code)
				assert.Contains(t, validationErr.Message, string(tt.status))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Message: "Test error message",
		Code:    "TEST_ERROR",
	}

	assert.Equal(t, "Test error message", err.Error())
}
