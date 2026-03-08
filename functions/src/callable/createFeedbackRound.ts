import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';

/**
 * Default feedback questions
 */
const DEFAULT_QUESTIONS = {
  q1: "What are this person's key strengths?",
  q2: "What areas could this person improve?",
  q3: "What specific behaviors or actions have you observed that stood out?",
  q4: "What advice would you give to help this person grow?",
};

/**
 * Callable Cloud Function to create a new feedback round.
 * Only admins can call this function.
 *
 * @param data - { subjectId: string, reviewerIds: string[], deadline: Date, questions?: object }
 * @param context - Firebase auth context
 */
export const createFeedbackRound = functions.https.onCall(async (data, context) => {
  // Verify user is authenticated
  if (!context.auth) {
    throw new functions.https.HttpsError(
      'unauthenticated',
      'User must be authenticated to perform this action.'
    );
  }

  const callerId = context.auth.uid;

  // Verify caller is an admin
  const callerDoc = await db.collection('users').doc(callerId).get();
  if (!callerDoc.exists || callerDoc.data()?.role !== 'admin') {
    throw new functions.https.HttpsError(
      'permission-denied',
      'Only admins can create feedback rounds.'
    );
  }

  // Validate input
  const { subjectId, reviewerIds, deadline, questions } = data;

  if (!subjectId || typeof subjectId !== 'string') {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'Subject ID is required and must be a string.'
    );
  }

  if (!Array.isArray(reviewerIds) || reviewerIds.length === 0) {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'At least one reviewer is required.'
    );
  }

  if (!deadline) {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'Deadline is required.'
    );
  }

  // Convert deadline to Date if it's a string or timestamp
  let deadlineDate: Date;
  if (typeof deadline === 'string') {
    deadlineDate = new Date(deadline);
  } else if (deadline._seconds) {
    deadlineDate = new Date(deadline._seconds * 1000);
  } else {
    deadlineDate = new Date(deadline);
  }

  // Validate deadline is in the future
  if (deadlineDate <= new Date()) {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'Deadline must be in the future.'
    );
  }

  // Verify subject exists
  const subjectDoc = await db.collection('users').doc(subjectId).get();
  if (!subjectDoc.exists) {
    throw new functions.https.HttpsError(
      'not-found',
      'Subject user not found.'
    );
  }
  const subjectData = subjectDoc.data();

  // Verify all reviewers exist
  const reviewerPromises = reviewerIds.map((id) => db.collection('users').doc(id).get());
  const reviewerDocs = await Promise.all(reviewerPromises);

  const invalidReviewers = reviewerDocs.filter((doc) => !doc.exists);
  if (invalidReviewers.length > 0) {
    throw new functions.https.HttpsError(
      'not-found',
      'One or more reviewer users not found.'
    );
  }

  // Get reviewer names
  const reviewerNames = reviewerDocs.map((doc) => doc.data()?.displayName || 'Unknown');

  // Subject cannot be a reviewer
  if (reviewerIds.includes(subjectId)) {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'Subject cannot be a reviewer for their own feedback.'
    );
  }

  // Create feedback round
  try {
    const roundRef = db.collection('feedbackRounds').doc();

    const roundData = {
      id: roundRef.id,
      subjectId,
      subjectName: subjectData?.displayName || 'Unknown',
      reviewerIds,
      reviewerNames,
      createdBy: callerId,
      createdAt: new Date(),
      deadline: deadlineDate,
      status: 'active',
      questions: questions || DEFAULT_QUESTIONS,
      submissionCount: 0,
      consolidatedAt: null,
      consolidatedBy: null,
      aiSummary: null,
      adminNotes: null,
      sharedAt: null,
      sharedBy: null,
    };

    await roundRef.set(roundData);

    functions.logger.info(
      `Feedback round created: ${roundRef.id} by ${callerId} for subject ${subjectId}`
    );

    return {
      success: true,
      roundId: roundRef.id,
      message: 'Feedback round created successfully.',
    };
  } catch (error) {
    functions.logger.error('Error creating feedback round:', error);
    throw new functions.https.HttpsError(
      'internal',
      'Failed to create feedback round. Please try again.'
    );
  }
});
