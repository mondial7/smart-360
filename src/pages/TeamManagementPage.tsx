/**
 * Team Management Page
 *
 * Admin page for managing team members and roles
 */

import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Button,
  Paper,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { TeamMemberList } from '../components/admin/TeamMemberList';

export const TeamManagementPage: React.FC = () => {
  const navigate = useNavigate();

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
          Team Management
        </Typography>
        <Typography variant="body1" color="text.secondary">
          View and manage team member roles. Admins can create feedback rounds and manage the team.
          Team members can submit and receive feedback.
        </Typography>
      </Box>

      <Paper sx={{ p: 0 }}>
        <TeamMemberList />
      </Paper>
    </Container>
  );
};
