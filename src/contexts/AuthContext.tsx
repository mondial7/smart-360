/**
 * Authentication Context
 *
 * Provides authentication state and Google Sign-In functionality
 */

import React, { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import type { User as FirebaseUser } from 'firebase/auth';
import {
  signInWithPopup,
  signOut as firebaseSignOut,
  onAuthStateChanged,
} from 'firebase/auth';
import { doc, getDoc, updateDoc, setDoc } from 'firebase/firestore';
import { auth, googleProvider, db } from '../config/firebase';
import type { User, UserFirestore } from '../types';

interface AuthContextType {
  currentUser: FirebaseUser | null;
  userProfile: User | null;
  loading: boolean;
  signInWithGoogle: () => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [currentUser, setCurrentUser] = useState<FirebaseUser | null>(null);
  const [userProfile, setUserProfile] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

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
   * Fetch user profile from Firestore
   */
  const fetchUserProfile = async (uid: string): Promise<User | null> => {
    try {
      const userDoc = await getDoc(doc(db, 'users', uid));
      if (userDoc.exists()) {
        const data = userDoc.data() as UserFirestore;
        return {
          ...data,
          createdAt: convertTimestamp(data.createdAt),
          lastLoginAt: convertTimestamp(data.lastLoginAt),
        };
      }
      return null;
    } catch (error) {
      console.error('Error fetching user profile:', error);
      return null;
    }
  };

  /**
   * Update last login timestamp
   */
  const updateLastLogin = async (uid: string) => {
    try {
      const userRef = doc(db, 'users', uid);
      const userDoc = await getDoc(userRef);

      // Only update if document exists (Cloud Function creates it for new users)
      if (userDoc.exists()) {
        await updateDoc(userRef, {
          lastLoginAt: new Date(),
        });
      }
      // If document doesn't exist, the Cloud Function will set lastLoginAt
    } catch (error) {
      console.error('Error updating last login:', error);
    }
  };

  /**
   * Create user profile in Firestore (fallback for emulator mode)
   */
  const createUserProfile = async (firebaseUser: FirebaseUser): Promise<User | null> => {
    try {
      // Check if this is the first user (for admin assignment)
      const configDoc = await getDoc(doc(db, 'users', '_config'));
      const isFirstUser = !configDoc.exists();

      const userData = {
        uid: firebaseUser.uid,
        email: firebaseUser.email || '',
        displayName: firebaseUser.displayName || firebaseUser.email?.split('@')[0] || 'User',
        photoURL: firebaseUser.photoURL || null,
        role: isFirstUser ? 'admin' : 'member',
        createdAt: new Date(),
        createdBy: null,
        lastLoginAt: new Date(),
        isActive: true,
      };

      await setDoc(doc(db, 'users', firebaseUser.uid), userData);

      // Mark that we've created at least one user
      if (isFirstUser) {
        await setDoc(doc(db, 'users', '_config'), { hasUsers: true });
      }

      return {
        ...userData,
        role: userData.role as User['role'],
        createdAt: userData.createdAt,
        lastLoginAt: userData.lastLoginAt,
      };
    } catch (error) {
      console.error('Error creating user profile:', error);
      return null;
    }
  };

  /**
   * Sign in with Google
   */
  const signInWithGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, googleProvider);
      // Check if user profile exists, create if not (for emulator mode)
      let profile = await fetchUserProfile(result.user.uid);
      if (!profile) {
        console.log('User profile not found, creating...');
        profile = await createUserProfile(result.user);
      }
      if (profile) {
        setUserProfile(profile);
      }
      // Update last login time
      await updateLastLogin(result.user.uid);
    } catch (error) {
      console.error('Error signing in with Google:', error);
      throw error;
    }
  };

  /**
   * Sign out
   */
  const signOut = async () => {
    try {
      await firebaseSignOut(auth);
      setUserProfile(null);
    } catch (error) {
      console.error('Error signing out:', error);
      throw error;
    }
  };

  /**
   * Listen to auth state changes
   */
  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, async (user) => {
      setCurrentUser(user);

      if (user) {
        // Fetch user profile from Firestore
        // Retry a few times if profile doesn't exist yet (Cloud Function might be creating it)
        let profile = await fetchUserProfile(user.uid);

        if (!profile) {
          // Wait and retry up to 3 times (total ~6 seconds)
          for (let i = 0; i < 3 && !profile; i++) {
            await new Promise(resolve => setTimeout(resolve, 2000));
            profile = await fetchUserProfile(user.uid);
          }
        }

        setUserProfile(profile);
      } else {
        setUserProfile(null);
      }

      setLoading(false);
    });

    return unsubscribe;
  }, []);

  const value: AuthContextType = {
    currentUser,
    userProfile,
    loading,
    signInWithGoogle,
    signOut,
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
};
