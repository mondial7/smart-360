/**
 * Feedback Rounds Page
 *
 * Admin page for managing feedback rounds
 */

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Button,
  Paper,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AddIcon from '@mui/icons-material/Add';
import { FeedbackRoundList } from '../components/admin/FeedbackRoundList';
import { FeedbackRoundCreator } from '../components/admin/FeedbackRoundCreator';

export const FeedbackRoundsPage: React.FC = () => {
  const navigate = useNavigate();
  const [creatorOpen, setCreatorOpen] = useState(false);

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

        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Box>
            <Typography variant="h4" component="h1" gutterBottom>
              Feedback Rounds
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Create and manage 360-degree feedback rounds for your team.
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreatorOpen(true)}
          >
            Create Round
          </Button>
        </Box>
      </Box>

      <Paper sx={{ p: 0 }}>
        <FeedbackRoundList />
      </Paper>

      <FeedbackRoundCreator
        open={creatorOpen}
        onClose={() => setCreatorOpen(false)}
      />
    </Container>
  );
};
