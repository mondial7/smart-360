/**
 * Smart 360 Feedback - Cloud Functions
 *
 * This file exports all Cloud Functions for the application.
 */

// Auth Triggers
export { onUserCreate } from './triggers/onUserCreate';

// Firestore Triggers
export { onFeedbackSubmit } from './triggers/onFeedbackSubmit';
export { onFeedbackRoundCreate } from './triggers/onFeedbackRoundCreate';
export { onFeedbackShared } from './triggers/onFeedbackShared';

// Scheduled Functions
export { sendDeadlineReminders } from './scheduled/sendDeadlineReminders';

// Callable Functions
export { assignUserRole } from './callable/assignUserRole';
export { createFeedbackRound } from './callable/createFeedbackRound';
export { consolidateFeedback } from './callable/consolidateFeedback';
export { shareFeedbackWithSubject } from './callable/shareFeedbackWithSubject';
