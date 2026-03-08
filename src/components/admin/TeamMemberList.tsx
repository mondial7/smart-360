/**
 * Team Member List Component
 *
 * Displays all team members with their roles and management options
 */

import React, { useState } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  Typography,
  Box,
  CircularProgress,
  Alert,
  Avatar,
} from '@mui/material';
import EditIcon from '@mui/icons-material/Edit';
import { useUsers } from '../../hooks/useUsers';
import { useAuth } from '../../contexts/AuthContext';
import { User } from '../../types';
import { TeamMemberForm } from './TeamMemberForm';

export const TeamMemberList: React.FC = () => {
  const { users, loading, error } = useUsers();
  const { currentUser } = useAuth();
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [formOpen, setFormOpen] = useState(false);

  const handleEditClick = (user: User) => {
    setSelectedUser(user);
    setFormOpen(true);
  };

  const handleFormClose = () => {
    setFormOpen(false);
    setSelectedUser(null);
  };

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
        Error loading team members: {error.message}
      </Alert>
    );
  }

  if (users.length === 0) {
    return (
      <Alert severity="info">
        No team members found. Users will appear here after signing in.
      </Alert>
    );
  }

  return (
    <>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>User</TableCell>
              <TableCell>Email</TableCell>
              <TableCell>Role</TableCell>
              <TableCell>Joined</TableCell>
              <TableCell>Last Login</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((user) => (
              <TableRow key={user.uid} hover>
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Avatar
                      src={user.photoURL || undefined}
                      alt={user.displayName}
                      sx={{ width: 32, height: 32 }}
                    >
                      {user.displayName.charAt(0).toUpperCase()}
                    </Avatar>
                    <Typography variant="body2">{user.displayName}</Typography>
                  </Box>
                </TableCell>
                <TableCell>
                  <Typography variant="body2">{user.email}</Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={user.role}
                    color={user.role === 'admin' ? 'primary' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2" color="text.secondary">
                    {user.createdAt.toLocaleDateString()}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" color="text.secondary">
                    {user.lastLoginAt.toLocaleDateString()}
                  </Typography>
                </TableCell>
                <TableCell align="center">
                  <IconButton
                    size="small"
                    onClick={() => handleEditClick(user)}
                    disabled={user.uid === currentUser?.uid}
                    title={
                      user.uid === currentUser?.uid
                        ? 'Cannot edit your own role'
                        : 'Edit user'
                    }
                  >
                    <EditIcon fontSize="small" />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {selectedUser && (
        <TeamMemberForm
          open={formOpen}
          user={selectedUser}
          onClose={handleFormClose}
        />
      )}
    </>
  );
};
