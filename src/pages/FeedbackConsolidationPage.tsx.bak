/**
 * Feedback Consolidation Page
 *
 * Page for consolidating and viewing feedback for a specific round
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
import { getFeedbackRoundById } from '../services/feedbackRoundService';
import type { FeedbackRound } from '../types';
import { FeedbackConsolidationView } from '../components/admin/FeedbackConsolidationView';

export const FeedbackConsolidationPage: React.FC = () => {
  const { roundId } = useParams<{ roundId: string }>();
  const navigate = useNavigate();

  const [round, setRound] = useState<FeedbackRound | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    const loadRound = async () => {
      if (!roundId) {
        setError('Invalid round ID');
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

        setRound(roundData);
        setLoading(false);
      } catch (err: any) {
        console.error('Error loading feedback round:', err);
        setError(err.message || 'Failed to load feedback round');
        setLoading(false);
      }
    };

    loadRound();
  }, [roundId]);

  if (loading) {
    return (
      <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  if (error || !round) {
    return (
      <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/admin/rounds')}
          sx={{ mb: 2 }}
        >
          Back to Rounds
        </Button>
        <Alert severity="error">{error || 'Failed to load feedback round'}</Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/admin/rounds')}
          sx={{ mb: 2 }}
        >
          Back to Rounds
        </Button>
        <Typography variant="h4" component="h1" gutterBottom>
          Consolidate Feedback
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Review and consolidate anonymous feedback using AI
        </Typography>
      </Box>

      <FeedbackConsolidationView round={round} />
    </Container>
  );
};
