/**
 * Admin Dashboard Component
 *
 * Main dashboard for admin users
 */

import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Paper,
  Grid,
  Card,
  CardContent,
  Button,
  CardActions,
} from '@mui/material';
import PeopleIcon from '@mui/icons-material/People';
import AssignmentIcon from '@mui/icons-material/Assignment';
import FeedbackIcon from '@mui/icons-material/Feedback';
import { useAuth } from '../../contexts/AuthContext';
import { useUsers } from '../../hooks/useUsers';

export const AdminDashboard: React.FC = () => {
  const { userProfile, signOut } = useAuth();
  const { users, loading } = useUsers();
  const navigate = useNavigate();

  const totalUsers = users.length;
  const adminCount = users.filter((u) => u.role === 'admin').length;
  const memberCount = users.filter((u) => u.role === 'member').length;

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Typography variant="h4" component="h1">
          Admin Dashboard
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
          Role: <strong>{userProfile?.role}</strong>
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Email: {userProfile?.email}
        </Typography>
      </Paper>

      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                <PeopleIcon color="primary" />
                <Typography variant="h6">Team Members</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Manage your team
              </Typography>
              <Typography variant="h3" sx={{ mb: 1 }}>
                {loading ? '-' : totalUsers}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {adminCount} admin{adminCount !== 1 ? 's' : ''}, {memberCount} member
                {memberCount !== 1 ? 's' : ''}
              </Typography>
            </CardContent>
            <CardActions>
              <Button size="small" onClick={() => navigate('/admin/team')}>
                Manage Team
              </Button>
            </CardActions>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                <AssignmentIcon color="primary" />
                <Typography variant="h6">Feedback Rounds</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Ongoing feedback rounds
              </Typography>
              <Typography variant="h3" sx={{ mb: 1 }}>
                0
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Coming in Phase 3
              </Typography>
            </CardContent>
            <CardActions>
              <Button size="small" disabled>
                View Rounds
              </Button>
            </CardActions>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                <FeedbackIcon color="primary" />
                <Typography variant="h6">Pending Feedback</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Awaiting consolidation
              </Typography>
              <Typography variant="h3" sx={{ mb: 1 }}>
                0
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Coming in Phase 4
              </Typography>
            </CardContent>
            <CardActions>
              <Button size="small" disabled>
                View Feedback
              </Button>
            </CardActions>
          </Card>
        </Grid>
      </Grid>

      <Box sx={{ mt: 4 }}>
        <Typography variant="body1" color="text.secondary">
          ✅ Phase 2 - Team member management is now available! Click "Manage Team" to view and edit user roles.
        </Typography>
      </Box>
    </Container>
  );
};
