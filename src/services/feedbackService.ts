/**
 * Feedback Service
 *
 * Handles all feedback submission operations
 */

import {
  collection,
  addDoc,
  getDocs,
  doc,
  getDoc,
  query,
  where,
  serverTimestamp,
  updateDoc,
} from 'firebase/firestore';
import { httpsCallable } from 'firebase/functions';
import { db, functions } from '../config/firebase';
import type {
  FeedbackSubmission,
  FeedbackSubmissionFirestore,
  FeedbackAnswers,
  ConsolidatedFeedback,
  ConsolidatedFeedbackFirestore,
  AISummary,
} from '../types';

/**
 * Convert Firestore timestamp to Date
 */
const convertTimestamp = (timestamp: any): Date => {
  if (timestamp instanceof Date) {
    return timestamp;
  }
  if (timestamp?.seconds) {
    return new Date(timestamp.seconds * 1000);
  }
  return new Date();
};

/**
 * Convert Firestore submission to FeedbackSubmission type
 */
const convertFirestoreSubmission = (
  data: FeedbackSubmissionFirestore
): FeedbackSubmission => {
  return {
    ...data,
    submittedAt: convertTimestamp(data.submittedAt),
    updatedAt: data.updatedAt ? convertTimestamp(data.updatedAt) : null,
  };
};

/**
 * Submit feedback for a round
 */
export const submitFeedback = async (
  roundId: string,
  reviewerId: string,
  subjectId: string,
  answers: FeedbackAnswers
): Promise<string> => {
  try {
    const submissionsRef = collection(db, 'feedbackSubmissions');

    const submissionData = {
      roundId,
      reviewerId,
      subjectId,
      answers,
      submittedAt: serverTimestamp(),
      updatedAt: null,
      isAnonymous: true,
    };

    const docRef = await addDoc(submissionsRef, submissionData);

    // Update the document with its own ID
    const submissionId = docRef.id;
    await updateDoc(docRef, { id: submissionId });

    return submissionId;
  } catch (error) {
    console.error('Error submitting feedback:', error);
    throw error;
  }
};

/**
 * Get all submissions for the current user (as reviewer)
 */
export const getMySubmissions = async (
  userId: string
): Promise<FeedbackSubmission[]> => {
  try {
    const submissionsRef = collection(db, 'feedbackSubmissions');
    const q = query(submissionsRef, where('reviewerId', '==', userId));
    const snapshot = await getDocs(q);

    return snapshot.docs.map((doc) => {
      const data = doc.data() as FeedbackSubmissionFirestore;
      return convertFirestoreSubmission(data);
    });
  } catch (error) {
    console.error('Error fetching my submissions:', error);
    throw error;
  }
};

/**
 * Get submission for a specific round by the current user
 */
export const getSubmissionForRound = async (
  roundId: string,
  reviewerId: string
): Promise<FeedbackSubmission | null> => {
  try {
    const submissionsRef = collection(db, 'feedbackSubmissions');
    const q = query(
      submissionsRef,
      where('roundId', '==', roundId),
      where('reviewerId', '==', reviewerId)
    );
    const snapshot = await getDocs(q);

    if (snapshot.empty) {
      return null;
    }

    const data = snapshot.docs[0].data() as FeedbackSubmissionFirestore;
    return convertFirestoreSubmission(data);
  } catch (error) {
    console.error('Error fetching submission for round:', error);
    throw error;
  }
};

/**
 * Check if user has submitted feedback for a round
 */
export const hasSubmittedFeedback = async (
  roundId: string,
  reviewerId: string
): Promise<boolean> => {
  const submission = await getSubmissionForRound(roundId, reviewerId);
  return submission !== null;
};

/**
 * Convert Firestore consolidated feedback to ConsolidatedFeedback type
 */
const convertFirestoreConsolidation = (
  data: ConsolidatedFeedbackFirestore
): ConsolidatedFeedback => {
  return {
    ...data,
    consolidatedAt: convertTimestamp(data.consolidatedAt),
    sharedAt: data.sharedAt ? convertTimestamp(data.sharedAt) : null,
  };
};

/**
 * Consolidate feedback for a round using AI (calls Cloud Function)
 */
export const consolidateFeedback = async (
  roundId: string
): Promise<{ consolidationId: string; aiSummary: AISummary }> => {
  try {
    const consolidateFunction = httpsCallable(functions, 'consolidateFeedback');
    const result = await consolidateFunction({ roundId });
    return result.data as { consolidationId: string; aiSummary: AISummary };
  } catch (error) {
    console.error('Error consolidating feedback:', error);
    throw error;
  }
};

/**
 * Get consolidated feedback by round ID
 */
export const getConsolidatedFeedbackByRoundId = async (
  roundId: string
): Promise<ConsolidatedFeedback | null> => {
  try {
    const consolidationsRef = collection(db, 'consolidatedFeedback');
    const q = query(consolidationsRef, where('roundId', '==', roundId));
    const snapshot = await getDocs(q);

    if (snapshot.empty) {
      return null;
    }

    const data = snapshot.docs[0].data() as ConsolidatedFeedbackFirestore;
    return convertFirestoreConsolidation(data);
  } catch (error) {
    console.error('Error fetching consolidated feedback:', error);
    throw error;
  }
};

/**
 * Get consolidated feedback by consolidation ID
 */
export const getConsolidatedFeedbackById = async (
  consolidationId: string
): Promise<ConsolidatedFeedback | null> => {
  try {
    const docRef = doc(db, 'consolidatedFeedback', consolidationId);
    const docSnap = await getDoc(docRef);

    if (!docSnap.exists()) {
      return null;
    }

    const data = docSnap.data() as ConsolidatedFeedbackFirestore;
    return convertFirestoreConsolidation(data);
  } catch (error) {
    console.error('Error fetching consolidated feedback by ID:', error);
    throw error;
  }
};

/**
 * Share consolidated feedback with subject (calls Cloud Function)
 */
export const shareFeedbackWithSubject = async (
  roundId: string,
  adminNotes?: string
): Promise<{ success: boolean; consolidationId: string }> => {
  try {
    const shareFunction = httpsCallable(functions, 'shareFeedbackWithSubject');
    const result = await shareFunction({ roundId, adminNotes });
    return result.data as { success: boolean; consolidationId: string };
  } catch (error) {
    console.error('Error sharing feedback with subject:', error);
    throw error;
  }
};
