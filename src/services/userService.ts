/**
 * User Service
 *
 * Handles all user-related operations with Firestore
 */

import {
  collection,
  getDocs,
  doc,
  getDoc,
  query,
  orderBy,
  updateDoc,
} from 'firebase/firestore';
import { httpsCallable } from 'firebase/functions';
import { db, functions } from '../config/firebase';
import { User, UserFirestore, UserRole } from '../types';

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

/**
 * Get all users from Firestore
 */
export const getAllUsers = async (): Promise<User[]> => {
  try {
    const usersRef = collection(db, 'users');
    const q = query(usersRef, orderBy('createdAt', 'desc'));
    const snapshot = await getDocs(q);

    return snapshot.docs.map((doc) => {
      const data = doc.data() as UserFirestore;
      return convertFirestoreUser(data);
    });
  } catch (error) {
    console.error('Error fetching users:', error);
    throw error;
  }
};

/**
 * Get a single user by UID
 */
export const getUserByUid = async (uid: string): Promise<User | null> => {
  try {
    const userDoc = await getDoc(doc(db, 'users', uid));
    if (userDoc.exists()) {
      const data = userDoc.data() as UserFirestore;
      return convertFirestoreUser(data);
    }
    return null;
  } catch (error) {
    console.error('Error fetching user:', error);
    throw error;
  }
};

/**
 * Update user role (calls Cloud Function for security)
 */
export const updateUserRole = async (
  uid: string,
  newRole: UserRole
): Promise<void> => {
  try {
    const assignUserRoleFunction = httpsCallable(functions, 'assignUserRole');
    await assignUserRoleFunction({ uid, role: newRole });
  } catch (error) {
    console.error('Error updating user role:', error);
    throw error;
  }
};

/**
 * Update user profile (non-role fields)
 */
export const updateUserProfile = async (
  uid: string,
  updates: Partial<Omit<User, 'uid' | 'role' | 'createdAt' | 'createdBy'>>
): Promise<void> => {
  try {
    const userRef = doc(db, 'users', uid);
    await updateDoc(userRef, updates);
  } catch (error) {
    console.error('Error updating user profile:', error);
    throw error;
  }
};
