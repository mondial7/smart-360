/**
 * Loading Skeleton Components
 *
 * Provides skeleton loaders for better UX during data loading
 */

import React from 'react';
import {
  Skeleton,
  Card,
  CardContent,
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
} from '@mui/material';

/**
 * Card list skeleton loader
 */
export const CardListSkeleton: React.FC<{ count?: number }> = ({ count = 3 }) => {
  return (
    <>
      {Array.from({ length: count }).map((_, index) => (
        <Card key={index} sx={{ mb: 2 }}>
          <CardContent>
            <Skeleton variant="text" width="60%" height={30} />
            <Skeleton variant="text" width="40%" sx={{ mt: 1 }} />
            <Box sx={{ display: 'flex', gap: 1, mt: 2 }}>
              <Skeleton variant="rectangular" width={80} height={24} />
              <Skeleton variant="rectangular" width={80} height={24} />
            </Box>
          </CardContent>
        </Card>
      ))}
    </>
  );
};

/**
 * Table skeleton loader
 */
export const TableSkeleton: React.FC<{ rows?: number; columns?: number }> = ({
  rows = 5,
  columns = 4,
}) => {
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            {Array.from({ length: columns }).map((_, index) => (
              <TableCell key={index}>
                <Skeleton variant="text" width="80%" />
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {Array.from({ length: rows }).map((_, rowIndex) => (
            <TableRow key={rowIndex}>
              {Array.from({ length: columns }).map((_, colIndex) => (
                <TableCell key={colIndex}>
                  <Skeleton variant="text" width="90%" />
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

/**
 * Dashboard card skeleton loader
 */
export const DashboardCardSkeleton: React.FC = () => {
  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Skeleton variant="circular" width={24} height={24} />
          <Skeleton variant="text" width="40%" height={30} />
        </Box>
        <Skeleton variant="text" width="60%" />
        <Skeleton variant="text" width="30%" height={48} sx={{ mt: 2 }} />
        <Skeleton variant="text" width="50%" />
      </CardContent>
    </Card>
  );
};

/**
 * Feedback form skeleton loader
 */
export const FeedbackFormSkeleton: React.FC = () => {
  return (
    <Paper sx={{ p: 3 }}>
      <Skeleton variant="text" width="60%" height={40} sx={{ mb: 3 }} />
      {Array.from({ length: 4 }).map((_, index) => (
        <Box key={index} sx={{ mb: 3 }}>
          <Skeleton variant="text" width="80%" />
          <Skeleton variant="rectangular" height={120} sx={{ mt: 1 }} />
        </Box>
      ))}
      <Skeleton variant="rectangular" width={200} height={48} />
    </Paper>
  );
};

/**
 * Feedback summary skeleton loader
 */
export const FeedbackSummarySkeleton: React.FC = () => {
  return (
    <Box>
      {Array.from({ length: 4 }).map((_, index) => (
        <Card key={index} sx={{ mb: 3 }}>
          <CardContent>
            <Skeleton variant="text" width="40%" height={32} sx={{ mb: 2 }} />
            <Skeleton variant="text" width="100%" />
            <Skeleton variant="text" width="100%" />
            <Skeleton variant="text" width="80%" />
          </CardContent>
        </Card>
      ))}
    </Box>
  );
};
