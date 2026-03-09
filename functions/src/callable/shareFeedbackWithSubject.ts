/**
 * Cloud Function: Share Feedback with Subject
 *
 * Allows admin to share consolidated feedback with the subject
 */

import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';

interface ShareFeedbackInput {
  roundId: string;
  adminNotes?: string;
}

interface ShareFeedbackOutput {
  success: boolean;
  consolidationId: string;
}

export const shareFeedbackWithSubject = functions.https.onCall(
  async (
    data: ShareFeedbackInput,
    context: functions.https.CallableContext
  ): Promise<ShareFeedbackOutput> => {
    // Verify authentication
    if (!context.auth) {
      throw new functions.https.HttpsError(
        'unauthenticated',
        'User must be authenticated to share feedback'
      );
    }

    const callerId = context.auth.uid;
    const { roundId, adminNotes } = data;

    // Validate input
    if (!roundId) {
      throw new functions.https.HttpsError(
        'invalid-argument',
        'roundId is required'
      );
    }

    try {
      // Verify caller is admin
      const callerDoc = await db.collection('users').doc(callerId).get();

      if (!callerDoc.exists) {
        throw new functions.https.HttpsError(
          'not-found',
          'User profile not found'
        );
      }

      const callerData = callerDoc.data();

      if (callerData?.role !== 'admin') {
        throw new functions.https.HttpsError(
          'permission-denied',
          'Only admins can share feedback'
        );
      }

      // Fetch feedback round
      const roundDoc = await db.collection('feedbackRounds').doc(roundId).get();

      if (!roundDoc.exists) {
        throw new functions.https.HttpsError(
          'not-found',
          'Feedback round not found'
        );
      }

      const roundData = roundDoc.data();

      if (!roundData) {
        throw new functions.https.HttpsError(
          'not-found',
          'Feedback round data is empty'
        );
      }

      // Check if already shared
      if (roundData.status === 'shared') {
        throw new functions.https.HttpsError(
          'failed-precondition',
          'This feedback has already been shared with the subject'
        );
      }

      // Verify round has been consolidated
      if (!roundData.consolidatedAt || !roundData.consolidationId) {
        throw new functions.https.HttpsError(
          'failed-precondition',
          'Feedback must be consolidated before sharing'
        );
      }

      const consolidationId = roundData.consolidationId;

      functions.logger.info('Sharing feedback with subject', {
        roundId,
        consolidationId,
        subjectId: roundData.subjectId,
        sharedBy: callerId,
      });

      // Update consolidated feedback document
      await db
        .collection('consolidatedFeedback')
        .doc(consolidationId)
        .update({
          sharedWithSubject: true,
          sharedAt: new Date(),
          adminNotes: adminNotes || null,
          adminApproved: true,
        });

      // Update feedback round status
      await db.collection('feedbackRounds').doc(roundId).update({
        status: 'shared',
        sharedAt: new Date(),
        sharedBy: callerId,
        adminNotes: adminNotes || null,
      });

      functions.logger.info('Successfully shared feedback with subject', {
        roundId,
        consolidationId,
        subjectId: roundData.subjectId,
      });

      return {
        success: true,
        consolidationId,
      };
    } catch (error: any) {
      functions.logger.error('Error sharing feedback with subject', {
        roundId,
        error: error.message,
        stack: error.stack,
      });

      // Re-throw HttpsError as-is
      if (error instanceof functions.https.HttpsError) {
        throw error;
      }

      // Wrap other errors
      throw new functions.https.HttpsError(
        'internal',
        `Failed to share feedback: ${error.message}`
      );
    }
  }
);
