/**
 * Team Member Form Component
 *
 * Dialog for editing user details and role
 */

import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  Box,
  Typography,
  Avatar,
  CircularProgress,
} from '@mui/material';
import { User, UserRole } from '../../types';
import { updateUserRole } from '../../services/userService';

interface TeamMemberFormProps {
  open: boolean;
  user: User;
  onClose: () => void;
}

export const TeamMemberForm: React.FC<TeamMemberFormProps> = ({
  open,
  user,
  onClose,
}) => {
  const [selectedRole, setSelectedRole] = useState<UserRole>(user.role);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState(false);

  const handleSubmit = async () => {
    if (selectedRole === user.role) {
      onClose();
      return;
    }

    try {
      setError('');
      setLoading(true);
      await updateUserRole(user.uid, selectedRole);
      setSuccess(true);
      setTimeout(() => {
        onClose();
        setSuccess(false);
      }, 1500);
    } catch (err: any) {
      console.error('Error updating user role:', err);
      setError(err.message || 'Failed to update user role. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setError('');
      setSuccess(false);
      setSelectedRole(user.role);
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Team Member</DialogTitle>
      <DialogContent>
        <Box sx={{ mb: 3, mt: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
            <Avatar
              src={user.photoURL || undefined}
              alt={user.displayName}
              sx={{ width: 56, height: 56 }}
            >
              {user.displayName.charAt(0).toUpperCase()}
            </Avatar>
            <Box>
              <Typography variant="h6">{user.displayName}</Typography>
              <Typography variant="body2" color="text.secondary">
                {user.email}
              </Typography>
            </Box>
          </Box>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          {success && (
            <Alert severity="success" sx={{ mb: 2 }}>
              User role updated successfully!
            </Alert>
          )}

          <FormControl fullWidth>
            <InputLabel id="role-select-label">Role</InputLabel>
            <Select
              labelId="role-select-label"
              id="role-select"
              value={selectedRole}
              label="Role"
              onChange={(e) => setSelectedRole(e.target.value as UserRole)}
              disabled={loading || success}
            >
              <MenuItem value="member">Team Member</MenuItem>
              <MenuItem value="admin">Admin</MenuItem>
            </Select>
          </FormControl>

          <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
            {selectedRole === 'admin'
              ? 'Admins can manage team members and create feedback rounds.'
              : 'Team members can submit and receive feedback.'}
          </Typography>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={loading || success}>
          Cancel
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={loading || success || selectedRole === user.role}
          startIcon={loading ? <CircularProgress size={16} /> : null}
        >
          {loading ? 'Saving...' : 'Save Changes'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};
