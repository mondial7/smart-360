/**
 * Scheduled Function: Send Deadline Reminders
 *
 * Runs daily at 9 AM to send reminders for upcoming deadlines
 * Sends reminders 3 days and 1 day before deadline
 */

import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';
import { sendEmail, createDeadlineReminderEmail } from '../services/emailService';

export const sendDeadlineReminders = functions.pubsub
  .schedule('0 9 * * *') // Run daily at 9 AM
  .timeZone('America/New_York') // Adjust to your timezone
  .onRun(async () => {
    functions.logger.info('Running deadline reminder job');

    try {
      // Calculate date ranges for 3 days and 1 day before deadline
      const now = new Date();
      const threeDaysFromNow = new Date(now);
      threeDaysFromNow.setDate(threeDaysFromNow.getDate() + 3);
      threeDaysFromNow.setHours(0, 0, 0, 0);

      const fourDaysFromNow = new Date(threeDaysFromNow);
      fourDaysFromNow.setDate(fourDaysFromNow.getDate() + 1);

      const oneDayFromNow = new Date(now);
      oneDayFromNow.setDate(oneDayFromNow.getDate() + 1);
      oneDayFromNow.setHours(0, 0, 0, 0);

      const twoDaysFromNow = new Date(oneDayFromNow);
      twoDaysFromNow.setDate(twoDaysFromNow.getDate() + 1);

      // Get active rounds with deadlines 3 days or 1 day away
      const roundsSnapshot = await db
        .collection('feedbackRounds')
        .where('status', '==', 'active')
        .get();

      let remindersSent = 0;

      for (const roundDoc of roundsSnapshot.docs) {
        const roundData = roundDoc.data();
        const deadline = roundData.deadline.toDate();
        deadline.setHours(0, 0, 0, 0);

        let daysRemaining = 0;
        let shouldSendReminder = false;

        // Check if deadline is 3 days away
        if (deadline.getTime() === threeDaysFromNow.getTime()) {
          daysRemaining = 3;
          shouldSendReminder = true;
        }
        // Check if deadline is 1 day away
        else if (deadline.getTime() === oneDayFromNow.getTime()) {
          daysRemaining = 1;
          shouldSendReminder = true;
        }

        if (!shouldSendReminder) continue;

        functions.logger.info('Sending deadline reminders for round', {
          roundId: roundDoc.id,
          daysRemaining,
        });

        // Get submissions for this round
        const submissionsSnapshot = await db
          .collection('feedbackSubmissions')
          .where('roundId', '==', roundDoc.id)
          .get();

        const submittedReviewerIds = new Set(
          submissionsSnapshot.docs.map((doc) => doc.data().reviewerId)
        );

        // Get reviewers who haven't submitted
        const pendingReviewerIds = roundData.reviewerIds.filter(
          (id: string) => !submittedReviewerIds.has(id)
        );

        if (pendingReviewerIds.length === 0) {
          functions.logger.info('All reviewers have submitted, skipping reminders', {
            roundId: roundDoc.id,
          });
          continue;
        }

        // Get app URL from config or use default
        const appUrl = functions.config().app?.url || 'http://localhost:5173';
        const feedbackUrl = `${appUrl}/feedback/submit/${roundDoc.id}`;

        // Send reminder to each pending reviewer
        for (const reviewerId of pendingReviewerIds) {
          const reviewerDoc = await db.collection('users').doc(reviewerId).get();
          const reviewerData = reviewerDoc.data();

          if (!reviewerData || !reviewerData.email) continue;

          const html = createDeadlineReminderEmail(
            reviewerData.displayName || reviewerData.email,
            roundData.subjectName,
            deadline,
            feedbackUrl,
            daysRemaining
          );

          await sendEmail({
            to: reviewerData.email,
            subject: `Reminder: Feedback Due in ${daysRemaining} Day${daysRemaining !== 1 ? 's' : ''}`,
            html,
          });

          remindersSent++;
        }
      }

      functions.logger.info('Deadline reminder job completed', { remindersSent });
    } catch (error: any) {
      functions.logger.error('Error in deadline reminder job', {
        error: error.message,
        stack: error.stack,
      });
    }
  });
