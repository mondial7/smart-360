/**
 * useFeedbackRounds Hook
 *
 * Custom hook for managing feedback rounds with real-time updates
 */

import { useState, useEffect } from 'react';
import { collection, query, orderBy, onSnapshot } from 'firebase/firestore';
import { db } from '../config/firebase';
import { FeedbackRound, FeedbackRoundFirestore } from '../types';

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

export const useFeedbackRounds = () => {
  const [rounds, setRounds] = useState<FeedbackRound[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const roundsRef = collection(db, 'feedbackRounds');
    const q = query(roundsRef, orderBy('createdAt', 'desc'));

    const unsubscribe = onSnapshot(
      q,
      (snapshot) => {
        const roundList = snapshot.docs.map((doc) => {
          const data = doc.data() as FeedbackRoundFirestore;
          return convertFirestoreRound(data);
        });
        setRounds(roundList);
        setLoading(false);
      },
      (err) => {
        console.error('Error fetching feedback rounds:', err);
        setError(err as Error);
        setLoading(false);
      }
    );

    return () => unsubscribe();
  }, []);

  return { rounds, loading, error };
};
