import * as admin from 'firebase-admin';

// Initialize Firebase Admin
admin.initializeApp();

export const db = admin.firestore();
export const auth = admin.auth();
export { admin };
