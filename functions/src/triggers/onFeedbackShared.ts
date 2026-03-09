/**
 * Trigger: On Feedback Shared
 *
 * Sends email notification to subject when feedback is shared with them
 */

import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';
import { sendEmail, createFeedbackSharedEmail } from '../services/emailService';

export const onFeedbackShared = functions.firestore
  .document('feedbackRounds/{roundId}')
  .onUpdate(async (change) => {
    const beforeData = change.before.data();
    const afterData = change.after.data();
    const roundId = change.after.id;

    // Check if status changed to 'shared'
    if (beforeData.status !== 'shared' && afterData.status === 'shared') {
      functions.logger.info('Feedback shared with subject', {
        roundId,
        subjectId: afterData.subjectId,
      });

      try {
        // Get subject details
        const subjectDoc = await db.collection('users').doc(afterData.subjectId).get();
        const subjectData = subjectDoc.data();

        if (!subjectData || !subjectData.email) {
          functions.logger.warn('Subject email not found', { subjectId: afterData.subjectId });
          return;
        }

        // Get app URL from config or use default
        const appUrl = functions.config().app?.url || 'http://localhost:5173';
        const feedbackUrl = `${appUrl}/my-feedback`;

        const html = createFeedbackSharedEmail(
          subjectData.displayName || subjectData.email,
          feedbackUrl
        );

        await sendEmail({
          to: subjectData.email,
          subject: 'Your 360 Feedback is Ready to View',
          html,
        });

        functions.logger.info('Sent feedback shared email', {
          roundId,
          subjectEmail: subjectData.email,
        });
      } catch (error: any) {
        functions.logger.error('Error sending feedback shared email', {
          roundId,
          error: error.message,
        });
        // Don't throw - we don't want to fail the update if email fails
      }
    }
  });
