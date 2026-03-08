/**
 * useUsers Hook
 *
 * Custom hook for managing user list with real-time updates
 */

import { useState, useEffect } from 'react';
import { collection, query, orderBy, onSnapshot } from 'firebase/firestore';
import { db } from '../config/firebase';
import { User, UserFirestore } from '../types';

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
 * Convert Firestore user document to User type
 */
const convertFirestoreUser = (data: UserFirestore): User => {
  return {
    ...data,
    createdAt: convertTimestamp(data.createdAt),
    lastLoginAt: convertTimestamp(data.lastLoginAt),
  };
};

export const useUsers = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const usersRef = collection(db, 'users');
    const q = query(usersRef, orderBy('createdAt', 'desc'));

    const unsubscribe = onSnapshot(
      q,
      (snapshot) => {
        const userList = snapshot.docs.map((doc) => {
          const data = doc.data() as UserFirestore;
          return convertFirestoreUser(data);
        });
        setUsers(userList);
        setLoading(false);
      },
      (err) => {
        console.error('Error fetching users:', err);
        setError(err as Error);
        setLoading(false);
      }
    );

    return () => unsubscribe();
  }, []);

  return { users, loading, error };
};
