/**
 * Pending Feedback List Component
 *
 * Displays feedback requests assigned to the current user
 */

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Paper,
  List,
  ListItem,
  ListItemText,
  ListItemButton,
  Chip,
  Typography,
  Box,
  CircularProgress,
  Alert,
  Divider,
} from '@mui/material';
import AssignmentIcon from '@mui/icons-material/Assignment';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { collection, query, where, onSnapshot } from 'firebase/firestore';
import { db } from '../../config/firebase';
import { useAuth } from '../../contexts/AuthContext';
import { FeedbackRound, FeedbackRoundFirestore } from '../../types';
import { hasSubmittedFeedback } from '../../services/feedbackService';

interface RoundWithStatus extends FeedbackRound {
  submitted: boolean;
}

export const PendingFeedbackList: React.FC = () => {
  const { currentUser } = useAuth();
  const navigate = useNavigate();
  const [rounds, setRounds] = useState<RoundWithStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!currentUser) return;

    const roundsRef = collection(db, 'feedbackRounds');
    const q = query(
      roundsRef,
      where('reviewerIds', 'array-contains', currentUser.uid),
      where('status', '==', 'active')
    );

    const unsubscribe = onSnapshot(
      q,
      async (snapshot) => {
        const roundsList: FeedbackRound[] = snapshot.docs.map((doc) => {
          const data = doc.data() as FeedbackRoundFirestore;
          return {
            ...data,
            createdAt:
              data.createdAt instanceof Date
                ? data.createdAt
                : new Date(data.createdAt.seconds * 1000),
            deadline:
              data.deadline instanceof Date
                ? data.deadline
                : new Date(data.deadline.seconds * 1000),
            consolidatedAt: data.consolidatedAt
              ? data.consolidatedAt instanceof Date
                ? data.consolidatedAt
                : new Date(data.consolidatedAt.seconds * 1000)
              : null,
            sharedAt: data.sharedAt
              ? data.sharedAt instanceof Date
                ? data.sharedAt
                : new Date(data.sharedAt.seconds * 1000)
              : null,
          };
        });

        // Check submission status for each round
        const roundsWithStatus = await Promise.all(
          roundsList.map(async (round) => {
            const submitted = await hasSubmittedFeedback(round.id, currentUser.uid);
            return { ...round, submitted };
          })
        );

        setRounds(roundsWithStatus);
        setLoading(false);
      },
      (err) => {
        console.error('Error fetching pending feedback:', err);
        setError(err as Error);
        setLoading(false);
      }
    );

    return () => unsubscribe();
  }, [currentUser]);

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error">
        Error loading feedback requests: {error.message}
      </Alert>
    );
  }

  if (rounds.length === 0) {
    return (
      <Alert severity="info">
        No pending feedback requests. You'll be notified when you're assigned to provide feedback.
      </Alert>
    );
  }

  const pendingRounds = rounds.filter((r) => !r.submitted);
  const completedRounds = rounds.filter((r) => r.submitted);

  return (
    <Paper>
      {pendingRounds.length > 0 && (
        <Box>
          <Box sx={{ p: 2, backgroundColor: 'primary.light', color: 'primary.contrastText' }}>
            <Typography variant="h6">Pending Feedback ({pendingRounds.length})</Typography>
          </Box>
          <List>
            {pendingRounds.map((round) => {
              const daysLeft = Math.ceil(
                (round.deadline.getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
              );
              const isOverdue = daysLeft < 0;

              return (
                <ListItem key={round.id} disablePadding>
                  <ListItemButton onClick={() => navigate(`/feedback/submit/${round.id}`)}>
                    <AssignmentIcon sx={{ mr: 2, color: 'primary.main' }} />
                    <ListItemText
                      primary={`Feedback for ${round.subjectName}`}
                      secondary={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                          <Typography variant="caption" color="text.secondary">
                            Due: {round.deadline.toLocaleDateString()}
                          </Typography>
                          <Chip
                            label={
                              isOverdue
                                ? `Overdue by ${Math.abs(daysLeft)} day${Math.abs(daysLeft) !== 1 ? 's' : ''}`
                                : daysLeft === 0
                                  ? 'Due today'
                                  : daysLeft === 1
                                    ? 'Due tomorrow'
                                    : `${daysLeft} days left`
                            }
                            size="small"
                            color={isOverdue ? 'error' : daysLeft <= 2 ? 'warning' : 'default'}
                          />
                        </Box>
                      }
                    />
                  </ListItemButton>
                </ListItem>
              );
            })}
          </List>
        </Box>
      )}

      {completedRounds.length > 0 && (
        <Box>
          {pendingRounds.length > 0 && <Divider />}
          <Box sx={{ p: 2, backgroundColor: 'success.light', color: 'success.contrastText' }}>
            <Typography variant="h6">Completed ({completedRounds.length})</Typography>
          </Box>
          <List>
            {completedRounds.map((round) => (
              <ListItem key={round.id}>
                <CheckCircleIcon sx={{ mr: 2, color: 'success.main' }} />
                <ListItemText
                  primary={`Feedback for ${round.subjectName}`}
                  secondary={`Submitted - Due: ${round.deadline.toLocaleDateString()}`}
                />
              </ListItem>
            ))}
          </List>
        </Box>
      )}
    </Paper>
  );
};
