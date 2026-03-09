/**
 * Trigger: On Feedback Round Created
 *
 * Sends email notifications to all reviewers when a new feedback round is created
 */

import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';
import { sendEmail, createFeedbackRequestEmail } from '../services/emailService';

export const onFeedbackRoundCreate = functions.firestore
  .document('feedbackRounds/{roundId}')
  .onCreate(async (snapshot) => {
    const roundData = snapshot.data();
    const roundId = snapshot.id;

    functions.logger.info('New feedback round created', {
      roundId,
      subjectId: roundData.subjectId,
      reviewerCount: roundData.reviewerIds.length,
    });

    try {
      // Get reviewer details
      const reviewerPromises = roundData.reviewerIds.map(async (reviewerId: string) => {
        const userDoc = await db.collection('users').doc(reviewerId).get();
        return { id: reviewerId, ...userDoc.data() };
      });

      const reviewers = await Promise.all(reviewerPromises);

      // Get app URL from config or use default
      const appUrl = functions.config().app?.url || 'http://localhost:5173';
      const feedbackUrl = `${appUrl}/feedback/submit/${roundId}`;

      // Send email to each reviewer
      const emailPromises = reviewers.map(async (reviewer: any) => {
        if (!reviewer.email) return;

        const html = createFeedbackRequestEmail(
          reviewer.displayName || reviewer.email,
          roundData.subjectName,
          roundData.deadline.toDate(),
          feedbackUrl
        );

        return sendEmail({
          to: reviewer.email,
          subject: `Feedback Request for ${roundData.subjectName}`,
          html,
        });
      });

      await Promise.all(emailPromises);

      functions.logger.info('Sent feedback request emails', {
        roundId,
        recipientCount: reviewers.length,
      });
    } catch (error: any) {
      functions.logger.error('Error sending feedback request emails', {
        roundId,
        error: error.message,
      });
      // Don't throw - we don't want to fail the round creation if emails fail
    }
  });
