/**
 * Cloud Function: Consolidate Feedback
 *
 * Consolidates anonymous feedback for a round using AI
 */

import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';
import { consolidateFeedbackWithAI, FeedbackSubmission, AISummary } from '../services/aiConsolidationService';

interface ConsolidateFeedbackInput {
  roundId: string;
}

interface ConsolidateFeedbackOutput {
  consolidationId: string;
  aiSummary: AISummary;
}

export const consolidateFeedback = functions.https.onCall(
  async (
    data: ConsolidateFeedbackInput,
    context: functions.https.CallableContext
  ): Promise<ConsolidateFeedbackOutput> => {
    // Verify authentication
    if (!context.auth) {
      throw new functions.https.HttpsError(
        'unauthenticated',
        'User must be authenticated to consolidate feedback'
      );
    }

    const callerId = context.auth.uid;
    const { roundId } = data;

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
          'Only admins can consolidate feedback'
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

      // Check if already consolidated
      if (roundData.consolidatedAt) {
        throw new functions.https.HttpsError(
          'failed-precondition',
          'This round has already been consolidated'
        );
      }

      // Verify round status
      if (roundData.status !== 'active' && roundData.status !== 'closed') {
        throw new functions.https.HttpsError(
          'failed-precondition',
          `Cannot consolidate round with status: ${roundData.status}`
        );
      }

      // Fetch all submissions for this round
      const submissionsSnapshot = await db
        .collection('feedbackSubmissions')
        .where('roundId', '==', roundId)
        .get();

      const submissions: FeedbackSubmission[] = submissionsSnapshot.docs.map((doc) => {
        const data = doc.data();
        return {
          id: doc.id,
          roundId: data.roundId,
          reviewerId: data.reviewerId,
          subjectId: data.subjectId,
          answers: data.answers,
          submittedAt: data.submittedAt?.toDate() || new Date(),
          isAnonymous: data.isAnonymous,
        };
      });

      // Validate minimum submissions
      if (submissions.length < 2) {
        throw new functions.https.HttpsError(
          'failed-precondition',
          `Need at least 2 submissions for consolidation. Current: ${submissions.length}`
        );
      }

      functions.logger.info('Consolidating feedback', {
        roundId,
        submissionCount: submissions.length,
        subjectId: roundData.subjectId,
      });

      // Call AI consolidation service
      const aiSummary = await consolidateFeedbackWithAI(
        submissions,
        roundData.questions
      );

      // Prepare raw feedback (anonymized - no reviewerIds exposed)
      const rawFeedback = submissions.map((submission) => ({
        submissionId: submission.id,
        answers: submission.answers,
      }));

      // Create consolidatedFeedback document
      const consolidationRef = db.collection('consolidatedFeedback').doc();

      const consolidationData = {
        id: consolidationRef.id,
        roundId,
        subjectId: roundData.subjectId,
        rawFeedback,
        aiSummary,
        adminNotes: null,
        adminApproved: false,
        consolidatedBy: callerId,
        consolidatedAt: new Date(),
        sharedWithSubject: false,
        sharedAt: null,
      };

      await consolidationRef.set(consolidationData);

      // Update feedback round with consolidation info
      await db.collection('feedbackRounds').doc(roundId).update({
        consolidatedAt: new Date(),
        consolidationId: consolidationRef.id,
        status: 'closed',
        aiSummary: aiSummary,
      });

      functions.logger.info('Successfully consolidated feedback', {
        roundId,
        consolidationId: consolidationRef.id,
      });

      return {
        consolidationId: consolidationRef.id,
        aiSummary,
      };
    } catch (error: any) {
      functions.logger.error('Error consolidating feedback', {
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
        `Failed to consolidate feedback: ${error.message}`
      );
    }
  }
);
