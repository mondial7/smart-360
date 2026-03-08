import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';

/**
 * Cloud Function triggered when a new user is created via Firebase Auth.
 * Creates a user document in Firestore with role assignment.
 * First user gets 'admin' role, subsequent users get 'member' role.
 */
export const onUserCreate = functions.auth.user().onCreate(async (user) => {
  try {
    // Check if this is the first user
    const usersSnapshot = await db.collection('users').limit(1).get();
    const isFirstUser = usersSnapshot.empty;

    // Create user document
    const userData = {
      uid: user.uid,
      email: user.email || '',
      displayName: user.displayName || user.email?.split('@')[0] || 'User',
      photoURL: user.photoURL || null,
      role: isFirstUser ? 'admin' : 'member',
      createdAt: new Date(),
      createdBy: null,
      lastLoginAt: new Date(),
      isActive: true,
    };

    await db.collection('users').doc(user.uid).set(userData);

    functions.logger.info(
      `User created: ${user.email} with role: ${userData.role}`,
      { uid: user.uid }
    );
  } catch (error) {
    functions.logger.error('Error creating user document:', error);
    throw error;
  }
});
