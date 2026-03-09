/**
 * Feedback Consolidation View Component
 *
 * Admin view for consolidating and reviewing feedback
 */

import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  CircularProgress,
  Alert,
  Tabs,
  Tab,
  Chip,
  Divider,
  Card,
  CardContent,
  LinearProgress,
} from '@mui/material';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import type { FeedbackRound, ConsolidatedFeedback } from '../../types';
import { consolidateFeedback, getConsolidatedFeedbackByRoundId } from '../../services/feedbackService';

interface FeedbackConsolidationViewProps {
  round: FeedbackRound;
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

const TabPanel: React.FC<TabPanelProps> = ({ children, value, index }) => {
  return (
    <div role="tabpanel" hidden={value !== index}>
      {value === index && <Box sx={{ py: 3 }}>{children}</Box>}
    </div>
  );
};

export const FeedbackConsolidationView: React.FC<FeedbackConsolidationViewProps> = ({ round }) => {
  const [consolidating, setConsolidating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [consolidation, setConsolidation] = useState<ConsolidatedFeedback | null>(null);
  const [tabValue, setTabValue] = useState(0);

  // Check if feedback is already consolidated
  useEffect(() => {
    const loadConsolidation = async () => {
      if (round.consolidatedAt) {
        try {
          const existing = await getConsolidatedFeedbackByRoundId(round.id);
          setConsolidation(existing);
        } catch (err: any) {
          console.error('Error loading consolidation:', err);
          setError(err.message || 'Failed to load consolidation');
        }
      }
      setLoading(false);
    };

    loadConsolidation();
  }, [round]);

  const handleConsolidate = async () => {
    setConsolidating(true);
    setError('');

    try {
      await consolidateFeedback(round.id);

      // Fetch the full consolidation
      const fullConsolidation = await getConsolidatedFeedbackByRoundId(round.id);
      setConsolidation(fullConsolidation);
    } catch (err: any) {
      console.error('Error consolidating feedback:', err);
      setError(err.message || 'Failed to consolidate feedback');
    } finally {
      setConsolidating(false);
    }
  };

  const isComplete = round.submissionCount === round.reviewerIds.length;
  const progress = (round.submissionCount / round.reviewerIds.length) * 100;

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      {/* Round Details Header */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h5" gutterBottom>
          Feedback Round for {round.subjectName}
        </Typography>

        <Box sx={{ mt: 2, mb: 2 }}>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Submission Progress
          </Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <LinearProgress
              variant="determinate"
              value={progress}
              sx={{ flexGrow: 1, height: 8, borderRadius: 4 }}
            />
            <Typography variant="body2" color="text.secondary">
              {round.submissionCount} / {round.reviewerIds.length}
            </Typography>
          </Box>
        </Box>

        <Box sx={{ display: 'flex', gap: 1, mt: 2 }}>
          <Chip
            label={round.status.toUpperCase()}
            color={round.status === 'closed' ? 'success' : 'primary'}
            size="small"
          />
          {isComplete && (
            <Chip
              icon={<CheckCircleIcon />}
              label="All Feedback Collected"
              color="success"
              size="small"
            />
          )}
        </Box>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Consolidation Status */}
      {!consolidation && (
        <Paper sx={{ p: 3 }}>
          {!isComplete && (
            <Alert severity="info" sx={{ mb: 2 }}>
              Waiting for all reviewers to submit feedback.
              {round.reviewerIds.length - round.submissionCount} submission(s) pending.
            </Alert>
          )}

          {isComplete && (
            <>
              <Alert severity="success" sx={{ mb: 2 }}>
                All feedback has been collected! You can now consolidate the feedback using AI.
              </Alert>

              <Button
                variant="contained"
                size="large"
                startIcon={<AutoAwesomeIcon />}
                onClick={handleConsolidate}
                disabled={consolidating}
                fullWidth
              >
                {consolidating ? 'Consolidating with AI...' : 'Consolidate Feedback with AI'}
              </Button>

              {consolidating && (
                <Box sx={{ mt: 2 }}>
                  <LinearProgress />
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 1, textAlign: 'center' }}>
                    This may take a moment...
                  </Typography>
                </Box>
              )}
            </>
          )}
        </Paper>
      )}

      {/* Consolidated Feedback View */}
      {consolidation && (
        <Paper sx={{ p: 0 }}>
          <Tabs value={tabValue} onChange={(_, newValue) => setTabValue(newValue)}>
            <Tab label="AI Summary" />
            <Tab label="Raw Feedback" />
          </Tabs>

          <TabPanel value={tabValue} index={0}>
            {/* AI Summary Tab */}
            <Box sx={{ px: 3 }}>
              <Alert severity="info" icon={<AutoAwesomeIcon />} sx={{ mb: 3 }}>
                This summary was generated by AI to consolidate all anonymous feedback responses.
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
                    {round.questions.q1}
                  </Typography>
                  <Typography variant="body2">
                    {consolidation.aiSummary.q1Summary}
                  </Typography>
                </CardContent>
              </Card>

              <Card sx={{ mb: 2 }}>
                <CardContent>
                  <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                    {round.questions.q2}
                  </Typography>
                  <Typography variant="body2">
                    {consolidation.aiSummary.q2Summary}
                  </Typography>
                </CardContent>
              </Card>

              <Card sx={{ mb: 2 }}>
                <CardContent>
                  <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                    {round.questions.q3}
                  </Typography>
                  <Typography variant="body2">
                    {consolidation.aiSummary.q3Summary}
                  </Typography>
                </CardContent>
              </Card>

              <Card sx={{ mb: 2 }}>
                <CardContent>
                  <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                    {round.questions.q4}
                  </Typography>
                  <Typography variant="body2">
                    {consolidation.aiSummary.q4Summary}
                  </Typography>
                </CardContent>
              </Card>
            </Box>
          </TabPanel>

          <TabPanel value={tabValue} index={1}>
            {/* Raw Feedback Tab */}
            <Box sx={{ px: 3 }}>
              <Alert severity="warning" sx={{ mb: 3 }}>
                This tab contains all raw feedback responses. Reviewer identities are anonymous.
              </Alert>

              {consolidation.rawFeedback.map((feedback, index) => (
                <Card key={feedback.submissionId} sx={{ mb: 2 }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      Reviewer {index + 1}
                    </Typography>
                    <Divider sx={{ mb: 2 }} />

                    <Box sx={{ mb: 2 }}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        {round.questions.q1}
                      </Typography>
                      <Typography variant="body2">
                        {feedback.answers.q1}
                      </Typography>
                    </Box>

                    <Box sx={{ mb: 2 }}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        {round.questions.q2}
                      </Typography>
                      <Typography variant="body2">
                        {feedback.answers.q2}
                      </Typography>
                    </Box>

                    <Box sx={{ mb: 2 }}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        {round.questions.q3}
                      </Typography>
                      <Typography variant="body2">
                        {feedback.answers.q3}
                      </Typography>
                    </Box>

                    <Box sx={{ mb: 0 }}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        {round.questions.q4}
                      </Typography>
                      <Typography variant="body2">
                        {feedback.answers.q4}
                      </Typography>
                    </Box>
                  </CardContent>
                </Card>
              ))}
            </Box>
          </TabPanel>
        </Paper>
      )}
    </Box>
  );
};
