/**
 * Main App Component
 *
 * Sets up routing and provides authentication context
 */

import React, { useMemo } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { CssBaseline, ThemeProvider, createTheme, Box, CircularProgress } from '@mui/material';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { SnackbarProvider } from './contexts/SnackbarContext';
import { ThemeModeProvider, useThemeMode } from './contexts/ThemeContext';
import { LoginPage } from './components/auth/LoginPage';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { AdminDashboard } from './components/admin/AdminDashboard';
import { MemberDashboard } from './components/member/MemberDashboard';
import { TeamManagementPage } from './pages/TeamManagementPage';
import { FeedbackRoundsPage } from './pages/FeedbackRoundsPage';
import { FeedbackSubmissionPage } from './pages/FeedbackSubmissionPage';
import { FeedbackConsolidationPage } from './pages/FeedbackConsolidationPage';
import { MyFeedbackPage } from './pages/MyFeedbackPage';

// Create theme based on mode
const useAppTheme = () => {
  const { mode } = useThemeMode();
  
  return useMemo(() => createTheme({
    palette: {
      mode,
      primary: {
        main: '#1976d2',
      },
      secondary: {
        main: '#dc004e',
      },
    },
  }), [mode]);
};

/**
 * Dashboard Router Component
 * Routes to appropriate dashboard based on user role
 */
const DashboardRouter: React.FC = () => {
  const { userProfile, loading } = useAuth();

  // Show loading while profile is being fetched
  // This prevents redirect loop when currentUser exists but userProfile is still loading
  if (loading) {
    return (
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '100vh',
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (!userProfile) {
    return <Navigate to="/" replace />;
  }

  return userProfile.role === 'admin' ? <AdminDashboard /> : <MemberDashboard />;
};

function AppContent() {
  const theme = useAppTheme();

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <SnackbarProvider>
        <AuthProvider>
          <Router>
          <Routes>
            {/* Public Route */}
            <Route path="/" element={<LoginPage />} />

            {/* Protected Routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardRouter />
                </ProtectedRoute>
              }
            />

            {/* Admin-only Routes */}
            <Route
              path="/admin/team"
              element={
                <ProtectedRoute requireAdmin={true}>
                  <TeamManagementPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/rounds"
              element={
                <ProtectedRoute requireAdmin={true}>
                  <FeedbackRoundsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/rounds/:roundId/consolidate"
              element={
                <ProtectedRoute requireAdmin={true}>
                  <FeedbackConsolidationPage />
                </ProtectedRoute>
              }
            />

            {/* Member Routes */}
            <Route
              path="/feedback/submit/:roundId"
              element={
                <ProtectedRoute>
                  <FeedbackSubmissionPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/my-feedback"
              element={
                <ProtectedRoute>
                  <MyFeedbackPage />
                </ProtectedRoute>
              }
            />

            {/* Catch all - redirect to home */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
          </Router>
        </AuthProvider>
      </SnackbarProvider>
    </ThemeProvider>
  );
}

function App() {
  return (
    <ThemeModeProvider>
      <AppContent />
    </ThemeModeProvider>
  );
}

export default App;
