import * as functions from 'firebase-functions';
import { db } from '../config/firebase-admin';

/**
 * Callable Cloud Function to assign or update a user's role.
 * Only admins can call this function.
 *
 * @param data - { uid: string, role: 'admin' | 'member' }
 * @param context - Firebase auth context
 */
export const assignUserRole = functions.https.onCall(async (data, context) => {
  // Verify user is authenticated
  if (!context.auth) {
    throw new functions.https.HttpsError(
      'unauthenticated',
      'User must be authenticated to perform this action.'
    );
  }

  const callerId = context.auth.uid;

  // Verify caller is an admin
  const callerDoc = await db.collection('users').doc(callerId).get();
  if (!callerDoc.exists || callerDoc.data()?.role !== 'admin') {
    throw new functions.https.HttpsError(
      'permission-denied',
      'Only admins can assign user roles.'
    );
  }

  // Validate input
  const { uid, role } = data;

  if (!uid || typeof uid !== 'string') {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'User ID (uid) is required and must be a string.'
    );
  }

  if (!role || !['admin', 'member'].includes(role)) {
    throw new functions.https.HttpsError(
      'invalid-argument',
      'Role must be either "admin" or "member".'
    );
  }

  // Prevent user from changing their own role
  if (uid === callerId) {
    throw new functions.https.HttpsError(
      'permission-denied',
      'You cannot change your own role.'
    );
  }

  // Verify target user exists
  const targetUserDoc = await db.collection('users').doc(uid).get();
  if (!targetUserDoc.exists) {
    throw new functions.https.HttpsError(
      'not-found',
      'Target user not found.'
    );
  }

  // Update user role
  try {
    await db.collection('users').doc(uid).update({
      role: role,
    });

    functions.logger.info(
      `User ${callerId} updated role for user ${uid} to ${role}`
    );

    return {
      success: true,
      message: `User role updated to ${role} successfully.`,
    };
  } catch (error) {
    functions.logger.error('Error updating user role:', error);
    throw new functions.https.HttpsError(
      'internal',
      'Failed to update user role. Please try again.'
    );
  }
});
