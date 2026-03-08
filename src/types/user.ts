/**
 * User role types
 */
export type UserRole = 'admin' | 'member';

/**
 * User interface matching Firestore user document
 */
export interface User {
  uid: string;
  email: string;
  displayName: string;
  photoURL: string | null;
  role: UserRole;
  createdAt: Date;
  createdBy: string | null;
  lastLoginAt: Date;
  isActive: boolean;
}

/**
 * User data from Firestore (before conversion)
 */
export interface UserFirestore {
  uid: string;
  email: string;
  displayName: string;
  photoURL: string | null;
  role: UserRole;
  createdAt: { seconds: number; nanoseconds: number } | Date;
  createdBy: string | null;
  lastLoginAt: { seconds: number; nanoseconds: number } | Date;
  isActive: boolean;
}
