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
} from 'firebase/firestore';
import { db } from '../config/firebase';
import {
  FeedbackSubmission,
  FeedbackSubmissionFirestore,
  FeedbackAnswers,
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
    await docRef.update({ id: submissionId });

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
