/**
 * Member Dashboard Component
 *
 * Main dashboard for team member users
 */

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Paper,
  Button,
  Card,
  CardContent,
} from '@mui/material';
import FeedbackIcon from '@mui/icons-material/Feedback';
import { useAuth } from '../../contexts/AuthContext';
import { PendingFeedbackList } from './PendingFeedbackList';
import { getSharedFeedbackForSubject } from '../../services/feedbackRoundService';
import { ThemeToggle } from '../common/ThemeToggle';
import type { FeedbackRound } from '../../types';

export const MemberDashboard: React.FC = () => {
  const { userProfile, currentUser, signOut } = useAuth();
  const navigate = useNavigate();
  const [sharedFeedback, setSharedFeedback] = useState<FeedbackRound[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadSharedFeedback = async () => {
      if (!currentUser) return;

      try {
        const rounds = await getSharedFeedbackForSubject(currentUser.uid);
        setSharedFeedback(rounds);
      } catch (error) {
        console.error('Error loading shared feedback:', error);
      } finally {
        setLoading(false);
      }
    };

    loadSharedFeedback();
  }, [currentUser]);

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Typography variant="h4" component="h1">
          My Dashboard
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <ThemeToggle />
          <Button variant="outlined" onClick={signOut}>
            Sign Out
          </Button>
        </Box>
      </Box>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Welcome, {userProfile?.displayName}!
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Email: {userProfile?.email}
        </Typography>
      </Paper>

      {/* My Feedback Section */}
      {!loading && sharedFeedback.length > 0 && (
        <Card sx={{ mb: 3, bgcolor: 'success.light', color: 'success.contrastText' }}>
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <FeedbackIcon fontSize="large" />
                <Box>
                  <Typography variant="h6">
                    You have {sharedFeedback.length} feedback round{sharedFeedback.length !== 1 ? 's' : ''} to review
                  </Typography>
                  <Typography variant="body2">
                    Your team has shared consolidated feedback with you
                  </Typography>
                </Box>
              </Box>
              <Button
                variant="contained"
                onClick={() => navigate('/my-feedback')}
                sx={{ bgcolor: 'white', color: 'success.main', '&:hover': { bgcolor: 'grey.100' } }}
              >
                View My Feedback
              </Button>
            </Box>
          </CardContent>
        </Card>
      )}

      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" gutterBottom>
          Feedback Requests
        </Typography>
        <PendingFeedbackList />
      </Box>
    </Container>
  );
};
