/**
 * Feedback Submission Page
 *
 * Page for submitting feedback for a specific round
 */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Button,
  CircularProgress,
  Alert,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { useAuth } from '../contexts/AuthContext';
import { getFeedbackRoundById } from '../services/feedbackRoundService';
import { getSubmissionForRound } from '../services/feedbackService';
import { FeedbackRound } from '../types';
import { FeedbackForm } from '../components/feedback/FeedbackForm';

export const FeedbackSubmissionPage: React.FC = () => {
  const { roundId } = useParams<{ roundId: string }>();
  const { currentUser } = useAuth();
  const navigate = useNavigate();

  const [round, setRound] = useState<FeedbackRound | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [alreadySubmitted, setAlreadySubmitted] = useState(false);

  useEffect(() => {
    const loadRound = async () => {
      if (!roundId || !currentUser) {
        setError('Invalid request');
        setLoading(false);
        return;
      }

      try {
        const roundData = await getFeedbackRoundById(roundId);

        if (!roundData) {
          setError('Feedback round not found');
          setLoading(false);
          return;
        }

        // Check if user is a reviewer for this round
        if (!roundData.reviewerIds.includes(currentUser.uid)) {
          setError('You are not authorized to provide feedback for this round');
          setLoading(false);
          return;
        }

        // Check if already submitted
        const submission = await getSubmissionForRound(roundId, currentUser.uid);
        if (submission) {
          setAlreadySubmitted(true);
        }

        setRound(roundData);
        setLoading(false);
      } catch (err: any) {
        console.error('Error loading feedback round:', err);
        setError(err.message || 'Failed to load feedback round');
        setLoading(false);
      }
    };

    loadRound();
  }, [roundId, currentUser]);

  if (loading) {
    return (
      <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  if (error || !round || !currentUser) {
    return (
      <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/dashboard')}
          sx={{ mb: 2 }}
        >
          Back to Dashboard
        </Button>
        <Alert severity="error">{error || 'Failed to load feedback round'}</Alert>
      </Container>
    );
  }

  if (alreadySubmitted) {
    return (
      <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/dashboard')}
          sx={{ mb: 2 }}>
          Back to Dashboard
        </Button>
        <Alert severity="info">
          You have already submitted feedback for this round. Thank you!
        </Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/dashboard')}
          sx={{ mb: 2 }}
        >
          Back to Dashboard
        </Button>
        <Typography variant="h4" component="h1" gutterBottom>
          Submit Feedback
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Provide anonymous feedback for your colleague. Be honest and constructive.
        </Typography>
      </Box>

      <FeedbackForm
        roundId={round.id}
        subjectId={round.subjectId}
        subjectName={round.subjectName}
        reviewerId={currentUser.uid}
        questions={round.questions}
        deadline={round.deadline}
      />
    </Container>
  );
};
