/**
 * Feedback Form Component
 *
 * Form for submitting feedback with 4 questions
 */

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
  Paper,
} from '@mui/material';
import { useForm, Controller } from 'react-hook-form';
import { FeedbackAnswers, FeedbackQuestions } from '../../types';
import { submitFeedback } from '../../services/feedbackService';

interface FeedbackFormProps {
  roundId: string;
  subjectId: string;
  subjectName: string;
  reviewerId: string;
  questions: FeedbackQuestions;
  deadline: Date;
}

export const FeedbackForm: React.FC<FeedbackFormProps> = ({
  roundId,
  subjectId,
  subjectName,
  reviewerId,
  questions,
  deadline,
}) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState(false);

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<FeedbackAnswers>({
    defaultValues: {
      q1: '',
      q2: '',
      q3: '',
      q4: '',
    },
  });

  const isOverdue = new Date() > deadline;

  const onSubmit = async (data: FeedbackAnswers) => {
    if (isOverdue) {
      setError('This feedback round has passed its deadline.');
      return;
    }

    try {
      setLoading(true);
      setError('');

      await submitFeedback(roundId, reviewerId, subjectId, data);

      setSuccess(true);
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err: any) {
      console.error('Error submitting feedback:', err);
      setError(err.message || 'Failed to submit feedback. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <Alert severity="success">
        Feedback submitted successfully! Redirecting to dashboard...
      </Alert>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Feedback for {subjectName}
        </Typography>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          Please answer the following questions honestly and constructively. Your responses will remain anonymous.
        </Typography>
        <Typography variant="caption" color="text.secondary">
          Deadline: {deadline.toLocaleDateString()} {deadline.toLocaleTimeString()}
        </Typography>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {isOverdue && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          This feedback round has passed its deadline. Submission may not be accepted.
        </Alert>
      )}

      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
        <Box>
          <Typography variant="subtitle1" gutterBottom fontWeight="medium">
            1. {questions.q1}
          </Typography>
          <Controller
            name="q1"
            control={control}
            rules={{ required: 'This field is required' }}
            render={({ field }) => (
              <TextField
                {...field}
                multiline
                rows={4}
                fullWidth
                placeholder="Your answer..."
                error={!!errors.q1}
                helperText={errors.q1?.message}
                disabled={loading || isOverdue}
              />
            )}
          />
        </Box>

        <Box>
          <Typography variant="subtitle1" gutterBottom fontWeight="medium">
            2. {questions.q2}
          </Typography>
          <Controller
            name="q2"
            control={control}
            rules={{ required: 'This field is required' }}
            render={({ field }) => (
              <TextField
                {...field}
                multiline
                rows={4}
                fullWidth
                placeholder="Your answer..."
                error={!!errors.q2}
                helperText={errors.q2?.message}
                disabled={loading || isOverdue}
              />
            )}
          />
        </Box>

        <Box>
          <Typography variant="subtitle1" gutterBottom fontWeight="medium">
            3. {questions.q3}
          </Typography>
          <Controller
            name="q3"
            control={control}
            rules={{ required: 'This field is required' }}
            render={({ field }) => (
              <TextField
                {...field}
                multiline
                rows={4}
                fullWidth
                placeholder="Your answer..."
                error={!!errors.q3}
                helperText={errors.q3?.message}
                disabled={loading || isOverdue}
              />
            )}
          />
        </Box>

        <Box>
          <Typography variant="subtitle1" gutterBottom fontWeight="medium">
            4. {questions.q4}
          </Typography>
          <Controller
            name="q4"
            control={control}
            rules={{ required: 'This field is required' }}
            render={({ field }) => (
              <TextField
                {...field}
                multiline
                rows={4}
                fullWidth
                placeholder="Your answer..."
                error={!!errors.q4}
                helperText={errors.q4?.message}
                disabled={loading || isOverdue}
              />
            )}
          />
        </Box>

        <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 2, mt: 2 }}>
          <Button
            variant="outlined"
            onClick={() => navigate('/dashboard')}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={loading || isOverdue}
            startIcon={loading ? <CircularProgress size={16} /> : null}
          >
            {loading ? 'Submitting...' : 'Submit Feedback'}
          </Button>
        </Box>
      </Box>
    </form>
  );
};
