/**
 * Feedback Round Service
 *
 * Handles all feedback round operations
 */

import {
  collection,
  getDocs,
  doc,
  getDoc,
  query,
  orderBy,
  where,
} from 'firebase/firestore';
import { httpsCallable } from 'firebase/functions';
import { db, functions } from '../config/firebase';
import {
  FeedbackRound,
  FeedbackRoundFirestore,
  CreateFeedbackRoundInput,
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
 * Convert Firestore feedback round to FeedbackRound type
 */
const convertFirestoreRound = (data: FeedbackRoundFirestore): FeedbackRound => {
  return {
    ...data,
    createdAt: convertTimestamp(data.createdAt),
    deadline: convertTimestamp(data.deadline),
    consolidatedAt: data.consolidatedAt ? convertTimestamp(data.consolidatedAt) : null,
    sharedAt: data.sharedAt ? convertTimestamp(data.sharedAt) : null,
  };
};

/**
 * Create a new feedback round (calls Cloud Function)
 */
export const createFeedbackRound = async (
  input: CreateFeedbackRoundInput
): Promise<{ roundId: string }> => {
  try {
    const createRoundFunction = httpsCallable(functions, 'createFeedbackRound');
    const result = await createRoundFunction(input);
    return result.data as { roundId: string };
  } catch (error) {
    console.error('Error creating feedback round:', error);
    throw error;
  }
};

/**
 * Get all feedback rounds
 */
export const getAllFeedbackRounds = async (): Promise<FeedbackRound[]> => {
  try {
    const roundsRef = collection(db, 'feedbackRounds');
    const q = query(roundsRef, orderBy('createdAt', 'desc'));
    const snapshot = await getDocs(q);

    return snapshot.docs.map((doc) => {
      const data = doc.data() as FeedbackRoundFirestore;
      return convertFirestoreRound(data);
    });
  } catch (error) {
    console.error('Error fetching feedback rounds:', error);
    throw error;
  }
};

/**
 * Get feedback round by ID
 */
export const getFeedbackRoundById = async (
  roundId: string
): Promise<FeedbackRound | null> => {
  try {
    const roundDoc = await getDoc(doc(db, 'feedbackRounds', roundId));
    if (roundDoc.exists()) {
      const data = roundDoc.data() as FeedbackRoundFirestore;
      return convertFirestoreRound(data);
    }
    return null;
  } catch (error) {
    console.error('Error fetching feedback round:', error);
    throw error;
  }
};

/**
 * Get feedback rounds for current user (as subject or reviewer)
 */
export const getMyFeedbackRounds = async (
  userId: string
): Promise<FeedbackRound[]> => {
  try {
    const roundsRef = collection(db, 'feedbackRounds');

    // Get rounds where user is subject
    const subjectQuery = query(
      roundsRef,
      where('subjectId', '==', userId),
      orderBy('createdAt', 'desc')
    );
    const subjectSnapshot = await getDocs(subjectQuery);

    // Get rounds where user is reviewer
    const reviewerQuery = query(
      roundsRef,
      where('reviewerIds', 'array-contains', userId),
      orderBy('createdAt', 'desc')
    );
    const reviewerSnapshot = await getDocs(reviewerQuery);

    // Combine and deduplicate
    const roundsMap = new Map<string, FeedbackRound>();

    subjectSnapshot.docs.forEach((doc) => {
      const data = doc.data() as FeedbackRoundFirestore;
      roundsMap.set(doc.id, convertFirestoreRound(data));
    });

    reviewerSnapshot.docs.forEach((doc) => {
      const data = doc.data() as FeedbackRoundFirestore;
      roundsMap.set(doc.id, convertFirestoreRound(data));
    });

    return Array.from(roundsMap.values()).sort(
      (a, b) => b.createdAt.getTime() - a.createdAt.getTime()
    );
  } catch (error) {
    console.error('Error fetching my feedback rounds:', error);
    throw error;
  }
};
