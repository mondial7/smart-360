/**
 * Member Dashboard Component
 *
 * Main dashboard for team member users
 */

import React from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  Button,
} from '@mui/material';
import { useAuth } from '../../contexts/AuthContext';
import { PendingFeedbackList } from './PendingFeedbackList';

export const MemberDashboard: React.FC = () => {
  const { userProfile, signOut } = useAuth();

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Typography variant="h4" component="h1">
          My Dashboard
        </Typography>
        <Button variant="outlined" onClick={signOut}>
          Sign Out
        </Button>
      </Box>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Welcome, {userProfile?.displayName}!
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Email: {userProfile?.email}
        </Typography>
      </Paper>

      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" gutterBottom>
          Feedback Requests
        </Typography>
        <PendingFeedbackList />
      </Box>

      <Box sx={{ mt: 4 }}>
        <Typography variant="body1" color="text.secondary">
          ✅ Phase 4 - You can now submit feedback! Click on any pending request to provide feedback.
        </Typography>
      </Box>
    </Container>
  );
};
