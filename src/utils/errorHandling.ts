/**
 * Error Handling Utilities
 *
 * Provides user-friendly error messages and error handling helpers
 */

/**
 * Extract user-friendly error message from Firebase error
 */
export const getErrorMessage = (error: any): string => {
  // Firebase Auth errors
  if (error.code) {
    switch (error.code) {
      case 'auth/popup-closed-by-user':
        return 'Sign-in cancelled. Please try again.';
      case 'auth/network-request-failed':
        return 'Network error. Please check your connection and try again.';
      case 'auth/user-disabled':
        return 'This account has been disabled. Please contact support.';
      case 'auth/too-many-requests':
        return 'Too many attempts. Please try again later.';

      // Firestore errors
      case 'permission-denied':
        return 'You don\'t have permission to perform this action.';
      case 'not-found':
        return 'The requested resource was not found.';
      case 'already-exists':
        return 'This resource already exists.';
      case 'resource-exhausted':
        return 'Too many requests. Please try again later.';
      case 'unavailable':
        return 'Service temporarily unavailable. Please try again.';

      // Functions errors
      case 'unauthenticated':
        return 'Please sign in to continue.';
      case 'failed-precondition':
        return error.message || 'Operation cannot be completed at this time.';
      case 'invalid-argument':
        return error.message || 'Invalid input. Please check your data.';
      case 'deadline-exceeded':
        return 'Request timed out. Please try again.';

      default:
        return error.message || 'An unexpected error occurred.';
    }
  }

  // Network errors
  if (error.message && error.message.includes('network')) {
    return 'Network error. Please check your connection and try again.';
  }

  // Generic error
  return error.message || 'An unexpected error occurred. Please try again.';
};

/**
 * Error logging helper
 */
export const logError = (context: string, error: any): void => {
  console.error(`[${context}]`, {
    message: error.message,
    code: error.code,
    stack: error.stack,
    details: error,
  });
};

/**
 * Retry wrapper for async operations
 */
export const retryOperation = async <T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  delayMs: number = 1000
): Promise<T> => {
  let lastError: any;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await operation();
    } catch (error: any) {
      lastError = error;

      // Don't retry on certain errors
      if (
        error.code === 'permission-denied' ||
        error.code === 'unauthenticated' ||
        error.code === 'invalid-argument' ||
        error.code === 'failed-precondition'
      ) {
        throw error;
      }

      if (attempt < maxRetries) {
        console.warn(`Attempt ${attempt} failed, retrying in ${delayMs}ms...`, error);
        await new Promise((resolve) => setTimeout(resolve, delayMs * attempt));
      }
    }
  }

  throw lastError;
};
