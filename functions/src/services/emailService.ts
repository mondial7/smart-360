/**
 * Email Service
 *
 * Handles sending emails for various notifications
 * Configure with SendGrid or SMTP in production
 */

import * as functions from 'firebase-functions';

export interface EmailOptions {
  to: string;
  subject: string;
  html: string;
}

/**
 * Send an email
 * In production, configure with SendGrid, AWS SES, or SMTP
 */
export const sendEmail = async (options: EmailOptions): Promise<void> => {
  // For now, just log the email (configure actual email service in production)
  functions.logger.info('Email notification', {
    to: options.to,
    subject: options.subject,
    preview: options.html.substring(0, 100),
  });

  // TODO: Implement actual email sending
  // Example with SendGrid:
  // const sgMail = require('@sendgrid/mail');
  // sgMail.setApiKey(functions.config().sendgrid.api_key);
  // await sgMail.send({
  //   to: options.to,
  //   from: 'noreply@your-domain.com',
  //   subject: options.subject,
  //   html: options.html,
  // });
};

/**
 * Create email template for feedback request
 */
export const createFeedbackRequestEmail = (
  recipientName: string,
  subjectName: string,
  deadline: Date,
  feedbackUrl: string
): string => {
  return `
    <!DOCTYPE html>
    <html>
    <head>
      <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1976d2; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; background: #1976d2; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 15px; }
        .deadline { background: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 15px 0; }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header">
          <h1>📝 Feedback Request</h1>
        </div>
        <div class="content">
          <p>Hi ${recipientName},</p>

          <p>You've been asked to provide feedback for <strong>${subjectName}</strong> as part of our 360-degree feedback process.</p>

          <div class="deadline">
            <strong>⏰ Deadline:</strong> ${deadline.toLocaleDateString('en-US', {
              weekday: 'long',
              year: 'numeric',
              month: 'long',
              day: 'numeric'
            })}
          </div>

          <p>Your honest and constructive feedback helps your colleagues grow professionally. Your responses are completely anonymous.</p>

          <a href="${feedbackUrl}" class="button">Submit Feedback</a>

          <p style="margin-top: 20px; font-size: 0.9em; color: #666;">
            This is an automated message from Smart 360 Feedback. Your feedback will remain anonymous.
          </p>
        </div>
      </div>
    </body>
    </html>
  `;
};

/**
 * Create email template for feedback shared notification
 */
export const createFeedbackSharedEmail = (
  recipientName: string,
  feedbackUrl: string
): string => {
  return `
    <!DOCTYPE html>
    <html>
    <head>
      <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4caf50; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; background: #4caf50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 15px; }
        .highlight { background: #e8f5e9; border-left: 4px solid #4caf50; padding: 10px; margin: 15px 0; }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header">
          <h1>🎯 Your Feedback is Ready</h1>
        </div>
        <div class="content">
          <p>Hi ${recipientName},</p>

          <div class="highlight">
            <strong>Good news!</strong> Your team has completed a feedback round and consolidated insights are now available for you to review.
          </div>

          <p>This feedback represents anonymous input from your colleagues, consolidated by AI to provide you with actionable insights for your professional development.</p>

          <a href="${feedbackUrl}" class="button">View My Feedback</a>

          <p style="margin-top: 20px; font-size: 0.9em; color: #666;">
            This is an automated message from Smart 360 Feedback.
          </p>
        </div>
      </div>
    </body>
    </html>
  `;
};

/**
 * Create email template for deadline reminder
 */
export const createDeadlineReminderEmail = (
  recipientName: string,
  subjectName: string,
  deadline: Date,
  feedbackUrl: string,
  daysRemaining: number
): string => {
  return `
    <!DOCTYPE html>
    <html>
    <head>
      <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #ff9800; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; background: #ff9800; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 15px; }
        .urgent { background: #fff3cd; border-left: 4px solid #ff9800; padding: 10px; margin: 15px 0; }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header">
          <h1>⏰ Feedback Deadline Reminder</h1>
        </div>
        <div class="content">
          <p>Hi ${recipientName},</p>

          <div class="urgent">
            <strong>Reminder:</strong> You have <strong>${daysRemaining} day${daysRemaining !== 1 ? 's' : ''}</strong> remaining to submit feedback for <strong>${subjectName}</strong>.
          </div>

          <p><strong>Deadline:</strong> ${deadline.toLocaleDateString('en-US', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric'
          })}</p>

          <p>Your feedback is valuable and helps your colleague grow. It only takes a few minutes to complete.</p>

          <a href="${feedbackUrl}" class="button">Submit Feedback Now</a>

          <p style="margin-top: 20px; font-size: 0.9em; color: #666;">
            This is an automated reminder from Smart 360 Feedback.
          </p>
        </div>
      </div>
    </body>
    </html>
  `;
};
