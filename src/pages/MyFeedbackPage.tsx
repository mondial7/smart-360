/**
 * My Feedback Page
 *
 * Page for team members to view feedback shared with them
 */

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Paper,
  Card,
  CardContent,
  CircularProgress,
  Alert,
  Chip,
  Button,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import FeedbackIcon from '@mui/icons-material/Feedback';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import { useAuth } from '../contexts/AuthContext';
import { getSharedFeedbackForSubject } from '../services/feedbackRoundService';
import { getConsolidatedFeedbackByRoundId } from '../services/feedbackService';
import type { FeedbackRound, ConsolidatedFeedback } from '../types';

export const MyFeedbackPage: React.FC = () => {
  const { currentUser } = useAuth();
  const navigate = useNavigate();

  const [rounds, setRounds] = useState<FeedbackRound[]>([]);
  const [selectedRound, setSelectedRound] = useState<FeedbackRound | null>(null);
  const [consolidation, setConsolidation] = useState<ConsolidatedFeedback | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    const loadSharedFeedback = async () => {
      if (!currentUser) return;

      try {
        const sharedRounds = await getSharedFeedbackForSubject(currentUser.uid);
        setRounds(sharedRounds);
        setLoading(false);
      } catch (err: any) {
        console.error('Error loading shared feedback:', err);
        setError(err.message || 'Failed to load feedback');
        setLoading(false);
      }
    };

    loadSharedFeedback();
  }, [currentUser]);

  const handleSelectRound = async (round: FeedbackRound) => {
    setSelectedRound(round);
    setError('');

    try {
      const consolidationData = await getConsolidatedFeedbackByRoundId(round.id);
      setConsolidation(consolidationData);
    } catch (err: any) {
      console.error('Error loading consolidation:', err);
      setError(err.message || 'Failed to load feedback details');
    }
  };

  if (loading) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/dashboard')}
          sx={{ mb: 2 }}
        >
          Back to Dashboard
        </Button>
        <Typography variant="h4" component="h1" gutterBottom>
          My Feedback
        </Typography>
        <Typography variant="body1" color="text.secondary">
          View consolidated feedback that has been shared with you
        </Typography>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {rounds.length === 0 ? (
        <Alert severity="info" icon={<FeedbackIcon />}>
          No feedback has been shared with you yet. When your team completes a feedback round,
          your admin will share the consolidated insights here.
        </Alert>
      ) : (
        <Box sx={{ display: 'flex', gap: 3 }}>
          {/* Feedback Rounds List */}
          <Box sx={{ flex: '0 0 300px' }}>
            <Typography variant="h6" gutterBottom>
              Feedback Rounds
            </Typography>
            {rounds.map((round) => (
              <Card
                key={round.id}
                sx={{
                  mb: 2,
                  cursor: 'pointer',
                  border: selectedRound?.id === round.id ? '2px solid' : '1px solid',
                  borderColor: selectedRound?.id === round.id ? 'primary.main' : 'divider',
                }}
                onClick={() => handleSelectRound(round)}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                    <Typography variant="body2" color="text.secondary">
                      {round.sharedAt?.toLocaleDateString()}
                    </Typography>
                    <Chip label="Shared" color="success" size="small" />
                  </Box>
                  <Typography variant="subtitle2">
                    Feedback from {round.reviewerIds.length} colleague{round.reviewerIds.length !== 1 ? 's' : ''}
                  </Typography>
                </CardContent>
              </Card>
            ))}
          </Box>

          {/* Feedback Details */}
          <Box sx={{ flex: 1 }}>
            {!selectedRound ? (
              <Alert severity="info">
                Select a feedback round from the list to view details
              </Alert>
            ) : !consolidation ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                <CircularProgress />
              </Box>
            ) : (
              <Paper sx={{ p: 3 }}>
                <Box sx={{ mb: 3 }}>
                  <Typography variant="h5" gutterBottom>
                    Feedback Round
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Shared on {selectedRound.sharedAt?.toLocaleDateString()}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Based on feedback from {selectedRound.reviewerIds.length} colleagues
                  </Typography>
                </Box>

                {selectedRound.adminNotes && (
                  <Alert severity="info" sx={{ mb: 3 }}>
                    <Typography variant="subtitle2" gutterBottom>
                      Note from your admin:
                    </Typography>
                    <Typography variant="body2">{selectedRound.adminNotes}</Typography>
                  </Alert>
                )}

                <Alert severity="info" icon={<AutoAwesomeIcon />} sx={{ mb: 3 }}>
                  This summary was generated by AI to consolidate anonymous feedback from your colleagues.
                </Alert>

                {/* Overview */}
                <Card sx={{ mb: 3 }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom color="primary">
                      Overview
                    </Typography>
                    <Typography variant="body1">
                      {consolidation.aiSummary.overview}
                    </Typography>
                  </CardContent>
                </Card>

                {/* Key Strengths */}
                <Card sx={{ mb: 3 }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom color="success.main">
                      Key Strengths
                    </Typography>
                    <Box component="ul" sx={{ pl: 2 }}>
                      {consolidation.aiSummary.keyStrengths.map((strength, index) => (
                        <Typography key={index} component="li" variant="body1" sx={{ mb: 1 }}>
                          {strength}
                        </Typography>
                      ))}
                    </Box>
                  </CardContent>
                </Card>

                {/* Areas for Improvement */}
                <Card sx={{ mb: 3 }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom color="warning.main">
                      Areas for Improvement
                    </Typography>
                    <Box component="ul" sx={{ pl: 2 }}>
                      {consolidation.aiSummary.areasForImprovement.map((area, index) => (
                        <Typography key={index} component="li" variant="body1" sx={{ mb: 1 }}>
                          {area}
                        </Typography>
                      ))}
                    </Box>
                  </CardContent>
                </Card>

                {/* Actionable Insights */}
                <Card sx={{ mb: 3 }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom color="info.main">
                      Actionable Insights
                    </Typography>
                    <Box component="ul" sx={{ pl: 2 }}>
                      {consolidation.aiSummary.actionableInsights.map((insight, index) => (
                        <Typography key={index} component="li" variant="body1" sx={{ mb: 1 }}>
                          {insight}
                        </Typography>
                      ))}
                    </Box>
                  </CardContent>
                </Card>

                {/* Per-Question Summaries */}
                <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
                  Question-by-Question Summary
                </Typography>

                <Card sx={{ mb: 2 }}>
                  <CardContent>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      {selectedRound.questions.q1}
                    </Typography>
                    <Typography variant="body2">
                      {consolidation.aiSummary.q1Summary}
                    </Typography>
                  </CardContent>
                </Card>

                <Card sx={{ mb: 2 }}>
                  <CardContent>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      {selectedRound.questions.q2}
                    </Typography>
                    <Typography variant="body2">
                      {consolidation.aiSummary.q2Summary}
                    </Typography>
                  </CardContent>
                </Card>

                <Card sx={{ mb: 2 }}>
                  <CardContent>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      {selectedRound.questions.q3}
                    </Typography>
                    <Typography variant="body2">
                      {consolidation.aiSummary.q3Summary}
                    </Typography>
                  </CardContent>
                </Card>

                <Card sx={{ mb: 2 }}>
                  <CardContent>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      {selectedRound.questions.q4}
                    </Typography>
                    <Typography variant="body2">
                      {consolidation.aiSummary.q4Summary}
                    </Typography>
                  </CardContent>
                </Card>
              </Paper>
            )}
          </Box>
        </Box>
      )}
    </Container>
  );
};
