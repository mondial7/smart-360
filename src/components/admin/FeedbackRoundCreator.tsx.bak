/**
 * Feedback Round Creator Component
 *
 * Multi-step form for creating new feedback rounds
 */

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Stepper,
  Step,
  StepLabel,
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Typography,
  Checkbox,
  FormGroup,
  FormControlLabel,
  Alert,
  CircularProgress,
  Chip,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { useUsers } from '../../hooks/useUsers';
import { useAuth } from '../../contexts/AuthContext';
import { createFeedbackRound } from '../../services/feedbackRoundService';
import { DEFAULT_QUESTIONS } from '../../config/feedbackQuestions';

interface FeedbackRoundCreatorProps {
  open: boolean;
  onClose: () => void;
}

const steps = ['Select Subject', 'Select Reviewers', 'Set Deadline', 'Review & Create'];

export const FeedbackRoundCreator: React.FC<FeedbackRoundCreatorProps> = ({
  open,
  onClose,
}) => {
  const { users } = useUsers();
  const { currentUser } = useAuth();
  const navigate = useNavigate();

  const [activeStep, setActiveStep] = useState(0);
  const [subjectId, setSubjectId] = useState('');
  const [reviewerIds, setReviewerIds] = useState<string[]>([]);
  const [deadline, setDeadline] = useState<Date | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const activeUsers = users.filter((u) => u.isActive);
  const availableSubjects = activeUsers;
  const availableReviewers = activeUsers.filter((u) => u.uid !== subjectId);

  const selectedSubject = users.find((u) => u.uid === subjectId);
  const selectedReviewers = users.filter((u) => reviewerIds.includes(u.uid));

  const handleNext = () => {
    setError('');

    // Validation
    if (activeStep === 0 && !subjectId) {
      setError('Please select a subject for feedback');
      return;
    }
    if (activeStep === 1 && reviewerIds.length === 0) {
      setError('Please select at least one reviewer');
      return;
    }
    if (activeStep === 2 && !deadline) {
      setError('Please select a deadline');
      return;
    }
    if (activeStep === 2 && deadline && deadline <= new Date()) {
      setError('Deadline must be in the future');
      return;
    }

    setActiveStep((prev) => prev + 1);
  };

  const handleBack = () => {
    setError('');
    setActiveStep((prev) => prev - 1);
  };

  const handleReviewerToggle = (userId: string) => {
    setReviewerIds((prev) =>
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId]
    );
  };

  const handleSubmit = async () => {
    if (!subjectId || reviewerIds.length === 0 || !deadline) {
      setError('Please complete all steps');
      return;
    }

    try {
      setLoading(true);
      setError('');

      const result = await createFeedbackRound({
        subjectId,
        reviewerIds,
        deadline,
      });

      // Success - close dialog and navigate
      onClose();
      navigate(`/admin/rounds`);
    } catch (err: any) {
      console.error('Error creating feedback round:', err);
      setError(err.message || 'Failed to create feedback round. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setActiveStep(0);
      setSubjectId('');
      setReviewerIds([]);
      setDeadline(null);
      setError('');
      onClose();
    }
  };

  const renderStepContent = () => {
    switch (activeStep) {
      case 0:
        // Select Subject
        return (
          <Box sx={{ minHeight: 200 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Select the person who will receive feedback
            </Typography>
            <FormControl fullWidth>
              <InputLabel>Subject</InputLabel>
              <Select
                value={subjectId}
                label="Subject"
                onChange={(e) => setSubjectId(e.target.value)}
              >
                {availableSubjects.map((user) => (
                  <MenuItem key={user.uid} value={user.uid}>
                    {user.displayName} ({user.email})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>
        );

      case 1:
        // Select Reviewers
        return (
          <Box sx={{ minHeight: 200 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Select team members who will provide feedback (excluding the subject)
            </Typography>
            <FormGroup>
              {availableReviewers.map((user) => (
                <FormControlLabel
                  key={user.uid}
                  control={
                    <Checkbox
                      checked={reviewerIds.includes(user.uid)}
                      onChange={() => handleReviewerToggle(user.uid)}
                      disabled={user.uid === currentUser?.uid}
                    />
                  }
                  label={`${user.displayName} (${user.email})`}
                />
              ))}
            </FormGroup>
            {availableReviewers.length === 0 && (
              <Alert severity="info">
                No other team members available to review this subject.
              </Alert>
            )}
          </Box>
        );

      case 2:
        // Set Deadline
        return (
          <Box sx={{ minHeight: 200 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Set a deadline for reviewers to submit their feedback
            </Typography>
            <LocalizationProvider dateAdapter={AdapterDateFns}>
              <DatePicker
                label="Deadline"
                value={deadline}
                onChange={(newValue) => setDeadline(newValue)}
                slotProps={{
                  textField: {
                    fullWidth: true,
                  },
                }}
              />
            </LocalizationProvider>
          </Box>
        );

      case 3:
        // Review & Create
        return (
          <Box sx={{ minHeight: 200 }}>
            <Typography variant="h6" gutterBottom>
              Review Feedback Round
            </Typography>

            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle2" color="text.secondary">
                Subject:
              </Typography>
              <Typography variant="body1">{selectedSubject?.displayName}</Typography>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle2" color="text.secondary">
                Reviewers ({reviewerIds.length}):
              </Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
                {selectedReviewers.map((reviewer) => (
                  <Chip key={reviewer.uid} label={reviewer.displayName} size="small" />
                ))}
              </Box>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle2" color="text.secondary">
                Deadline:
              </Typography>
              <Typography variant="body1">
                {deadline?.toLocaleDateString()} {deadline?.toLocaleTimeString()}
              </Typography>
            </Box>

            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                Feedback Questions:
              </Typography>
              {Object.entries(DEFAULT_QUESTIONS).map(([key, question]) => (
                <Typography key={key} variant="body2" sx={{ mb: 1 }}>
                  {key.toUpperCase()}. {question}
                </Typography>
              ))}
            </Box>
          </Box>
        );

      default:
        return null;
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>Create Feedback Round</DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 2 }}>
          <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          {renderStepContent()}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={loading}>
          Cancel
        </Button>
        {activeStep > 0 && (
          <Button onClick={handleBack} disabled={loading}>
            Back
          </Button>
        )}
        {activeStep < steps.length - 1 ? (
          <Button onClick={handleNext} variant="contained" disabled={loading}>
            Next
          </Button>
        ) : (
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={loading}
            startIcon={loading ? <CircularProgress size={16} /> : null}
          >
            {loading ? 'Creating...' : 'Create Round'}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};
