/**
 * Main App Component
 *
 * Sets up routing and provides authentication context
 */

import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { SnackbarProvider } from './contexts/SnackbarContext';
import { LoginPage } from './components/auth/LoginPage';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { AdminDashboard } from './components/admin/AdminDashboard';
import { MemberDashboard } from './components/member/MemberDashboard';
import { TeamManagementPage } from './pages/TeamManagementPage';
import { FeedbackRoundsPage } from './pages/FeedbackRoundsPage';
import { FeedbackSubmissionPage } from './pages/FeedbackSubmissionPage';
import { FeedbackConsolidationPage } from './pages/FeedbackConsolidationPage';
import { MyFeedbackPage } from './pages/MyFeedbackPage';

// Create Material-UI theme
const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

/**
 * Dashboard Router Component
 * Routes to appropriate dashboard based on user role
 */
const DashboardRouter: React.FC = () => {
  const { userProfile } = useAuth();

  if (!userProfile) {
    return <Navigate to="/" replace />;
  }

  return userProfile.role === 'admin' ? <AdminDashboard /> : <MemberDashboard />;
};

function App() {
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

export default App;
