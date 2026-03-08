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
  Grid,
  Card,
  CardContent,
  Button,
} from '@mui/material';
import { useAuth } from '../../contexts/AuthContext';

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

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6">Pending Feedback Requests</Typography>
              <Typography variant="body2" color="text.secondary">
                Feedback you need to provide
              </Typography>
              <Typography variant="h4" sx={{ mt: 2 }}>
                -
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6">Received Feedback</Typography>
              <Typography variant="body2" color="text.secondary">
                Your consolidated feedback
              </Typography>
              <Typography variant="h4" sx={{ mt: 2 }}>
                -
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box sx={{ mt: 4 }}>
        <Typography variant="body1" color="text.secondary">
          Phase 1 Complete - Authentication working! Next phases will add feedback submission and viewing.
        </Typography>
      </Box>
    </Container>
  );
};
