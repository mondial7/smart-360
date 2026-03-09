import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';

/**
 * Cloud Function triggered when a feedback submission is created.
 * Increments the submission count on the feedback round.
 */
export const onFeedbackSubmit = functions.firestore
  .document('feedbackSubmissions/{submissionId}')
  .onCreate(async (snapshot, context) => {
    try {
      const submissionData = snapshot.data();
      const roundId = submissionData.roundId;

      if (!roundId) {
        functions.logger.error('Submission missing roundId', {
          submissionId: context.params.submissionId,
        });
        return;
      }

      // Increment the submission count on the feedback round
      const roundRef = db.collection('feedbackRounds').doc(roundId);

      await roundRef.update({
        submissionCount: (submissionData.submissionCount || 0) + 1,
      });

      // Get the updated round to check if all submissions are in
      const roundDoc = await roundRef.get();
      const roundData = roundDoc.data();

      if (roundData) {
        const totalReviewers = roundData.reviewerIds?.length || 0;
        const currentSubmissions = (roundData.submissionCount || 0) + 1;

        functions.logger.info(
          `Feedback submitted for round ${roundId}. Progress: ${currentSubmissions}/${totalReviewers}`,
          {
            roundId,
            submissionId: context.params.submissionId,
            reviewerId: submissionData.reviewerId,
          }
        );

        // Optional: Notify admin when all feedback is collected
        if (currentSubmissions === totalReviewers) {
          functions.logger.info(
            `All feedback collected for round ${roundId}. Ready for consolidation.`,
            { roundId }
          );
          // Future: Create notification for admin
        }
      }
    } catch (error) {
      functions.logger.error('Error processing feedback submission:', error);
    }
  });
