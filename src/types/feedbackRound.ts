/**
 * Feedback Round Types
 */

export type FeedbackRoundStatus = 'draft' | 'active' | 'closed' | 'shared';

/**
 * Feedback questions structure
 */
export interface FeedbackQuestions {
  q1: string;
  q2: string;
  q3: string;
  q4: string;
}

/**
 * Feedback round interface
 */
export interface FeedbackRound {
  id: string;
  subjectId: string;
  subjectName: string;
  reviewerIds: string[];
  reviewerNames: string[];
  createdBy: string;
  createdAt: Date;
  deadline: Date;
  status: FeedbackRoundStatus;
  questions: FeedbackQuestions;
  submissionCount: number;
  consolidatedAt: Date | null;
  consolidatedBy: string | null;
  aiSummary: string | null;
  adminNotes: string | null;
  sharedAt: Date | null;
  sharedBy: string | null;
}

/**
 * Firestore feedback round (before conversion)
 */
export interface FeedbackRoundFirestore {
  id: string;
  subjectId: string;
  subjectName: string;
  reviewerIds: string[];
  reviewerNames: string[];
  createdBy: string;
  createdAt: { seconds: number; nanoseconds: number } | Date;
  deadline: { seconds: number; nanoseconds: number } | Date;
  status: FeedbackRoundStatus;
  questions: FeedbackQuestions;
  submissionCount: number;
  consolidatedAt: { seconds: number; nanoseconds: number } | Date | null;
  consolidatedBy: string | null;
  aiSummary: string | null;
  adminNotes: string | null;
  sharedAt: { seconds: number; nanoseconds: number } | Date | null;
  sharedBy: string | null;
}

/**
 * Input for creating a new feedback round
 */
export interface CreateFeedbackRoundInput {
  subjectId: string;
  reviewerIds: string[];
  deadline: Date;
  questions?: FeedbackQuestions; // Optional override
}
