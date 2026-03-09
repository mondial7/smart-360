/**
 * Theme Toggle Component
 *
 * Button to toggle between light and dark themes
 */

import React from 'react';
import { IconButton, Tooltip } from '@mui/material';
import { Brightness4, Brightness7 } from '@mui/icons-material';
import { useThemeMode } from '../../contexts/ThemeContext';

interface ThemeToggleProps {
  size?: 'small' | 'medium' | 'large';
  color?: 'inherit' | 'default' | 'primary' | 'secondary';
}

export const ThemeToggle: React.FC<ThemeToggleProps> = ({
  size = 'medium',
  color = 'inherit',
}) => {
  const { mode, toggleTheme } = useThemeMode();
  const isDark = mode === 'dark';

  return (
    <Tooltip title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}>
      <IconButton onClick={toggleTheme} color={color} size={size}>
        {isDark ? <Brightness7 /> : <Brightness4 />}
      </IconButton>
    </Tooltip>
  );
};
