# Smart 360 Feedback

A serverless web application for collecting, consolidating, and delivering anonymous 360-degree feedback.

## Features

- **User Authentication**: Google Sign-In only (no signup screens)
- **Role-Based Access**: Admin and Team Member roles
- **Feedback Rounds**: Admins create rounds with selected reviewers and deadlines
- **Anonymous Feedback**: Team members submit feedback via 4-question template
- **AI Consolidation**: OpenAI-powered feedback summarization
- **Secure Delivery**: Share consolidated feedback with subjects

## Tech Stack

- **Frontend**: React 18 + TypeScript + Vite + Material-UI
- **Backend**: Firebase (Firestore, Auth, Cloud Functions)
- **AI**: OpenAI GPT-4o-mini
- **Deployment**: Firebase Hosting

## Project Status

✅ **Phase 1 Complete**: Project Foundation & Authentication
- React app initialized with TypeScript
- Firebase project created and configured
- Google Sign-In authentication implemented
- User role management (admin/member)
- Protected routes and dashboards

🚧 **Next Phases**:
- Phase 2: User Role Management
- Phase 3: Feedback Rounds
- Phase 4: Feedback Submission
- Phase 5: AI Consolidation
- Phase 6: Feedback Delivery
- Phase 7: Polish & Production

## Prerequisites

- Node.js 18+ and npm
- Firebase CLI (`npm install -g firebase-tools`)
- Google account for Firebase
- Firebase Blaze plan (for Cloud Functions deployment)

## Setup Instructions

### 1. Clone and Install Dependencies

```bash
# Install frontend dependencies
npm install

# Install Cloud Functions dependencies
cd functions
npm install
cd ..
```

### 2. Firebase Configuration

The Firebase project is already created: `smart-360-feedback`

**Enable Google Authentication:**

1. Go to [Firebase Console](https://console.firebase.google.com/project/smart-360-feedback/authentication/providers)
2. Click on "Authentication" in the left sidebar
3. Go to "Sign-in method" tab
4. Click on "Google" provider
5. Toggle "Enable"
6. Add your email to "Project support email"
7. Click "Save"

### 3. Upgrade to Blaze Plan (Required for Cloud Functions)

To deploy Cloud Functions, you need to upgrade to the Blaze (pay-as-you-go) plan:

1. Visit: https://console.firebase.google.com/project/smart-360-feedback/usage/details
2. Click "Upgrade" and follow the instructions
3. Add a billing account (Note: Firebase has generous free tier)

**Note**: For local development, you can use Firebase emulators without upgrading.

### 4. Environment Variables

The `.env.local` file has been created with your Firebase configuration:

```
VITE_FIREBASE_API_KEY=AIzaSyBLQaHyPD64mqdHf5teOE_0SRw-j0R--Pg
VITE_FIREBASE_AUTH_DOMAIN=smart-360-feedback.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=smart-360-feedback
VITE_FIREBASE_STORAGE_BUCKET=smart-360-feedback.firebasestorage.app
VITE_FIREBASE_MESSAGING_SENDER_ID=710920208749
VITE_FIREBASE_APP_ID=1:710920208749:web:216eef44db4c8b02321a64
```

⚠️ **Important**: Never commit `.env.local` to version control (it's already in `.gitignore`)

## Development

### Option 1: Local Development with Firebase Emulators (Recommended)

```bash
# Terminal 1: Start Firebase emulators
firebase emulators:start

# Terminal 2: Start React dev server
npm run dev
```

The app will be available at `http://localhost:5173`
Firebase Emulator UI at `http://localhost:4000`

**Benefits:**
- No need for Blaze plan
- Faster development
- Offline development
- No costs for testing

### Option 2: Development with Production Firebase

```bash
# Make sure you've upgraded to Blaze plan and enabled Google Auth

# Deploy Cloud Functions and Firestore rules
firebase deploy

# Start React dev server
npm run dev
```

## Deployment

### Deploy Everything

```bash
# Build the React app
npm run build

# Deploy to Firebase
firebase deploy
```

### Deploy Specific Services

```bash
# Deploy only Firestore rules
firebase deploy --only firestore:rules

# Deploy only Cloud Functions
firebase deploy --only functions

# Deploy only hosting
firebase deploy --only hosting
```

## Project Structure

```
smart-360/
├── src/                          # React frontend
│   ├── config/
│   │   └── firebase.ts           # Firebase client config
│   ├── contexts/
│   │   └── AuthContext.tsx       # Authentication context
│   ├── components/
│   │   ├── auth/
│   │   │   ├── LoginPage.tsx
│   │   │   └── ProtectedRoute.tsx
│   │   ├── admin/
│   │   │   └── AdminDashboard.tsx
│   │   └── member/
│   │       └── MemberDashboard.tsx
│   ├── types/
│   │   └── user.ts               # TypeScript types
│   └── App.tsx                   # Main app with routing
│
├── functions/                    # Cloud Functions
│   ├── src/
│   │   ├── triggers/
│   │   │   └── onUserCreate.ts   # Auto-create user docs
│   │   └── index.ts
│   └── package.json
│
├── firestore.rules               # Firestore security rules
├── firestore.indexes.json        # Firestore indexes
├── firebase.json                 # Firebase configuration
└── .firebaserc                   # Firebase project alias
```

## User Roles

### Admin
- First user to sign in automatically becomes admin
- Can manage team members
- Can create feedback rounds
- Can consolidate and share feedback

### Team Member
- All subsequent users are team members by default
- Can submit feedback
- Can view received feedback
- Admins can promote members to admin

## How It Works

1. **Sign In**: User signs in with Google
2. **Auto-Setup**: Cloud Function creates user document with role
3. **First User**: Automatically assigned 'admin' role
4. **Dashboard**: Redirected to role-appropriate dashboard
5. **Future Phases**: Will add feedback rounds, submissions, and AI consolidation

## Testing the Authentication Flow

1. Start the development server (with emulators or production)
2. Navigate to `http://localhost:5173`
3. Click "Sign in with Google"
4. Authorize the application
5. You should be redirected to the Admin Dashboard (as first user)
6. Sign out and test with another account (will be assigned 'member' role)

## Firestore Collections

### users
```typescript
{
  uid: string
  email: string
  displayName: string
  photoURL: string | null
  role: 'admin' | 'member'
  createdAt: Date
  createdBy: string | null
  lastLoginAt: Date
  isActive: boolean
}
```

## Security

- **Authentication**: Google Sign-In only
- **Firestore Rules**: Role-based access control
- **Cloud Functions**: Server-side validation
- **Anonymity**: Reviewer IDs never exposed to subjects

## Troubleshooting

### "Missing required API" errors
- Upgrade to Blaze plan
- Or use Firebase emulators for local development

### Google Sign-In not working
- Make sure Google provider is enabled in Firebase Console
- Check that authorized domains include your localhost and production domain

### Emulator connection errors
- Make sure emulators are running: `firebase emulators:start`
- Check ports are not in use (4000, 5001, 8080, 9099)

### Build errors
- Clear node_modules: `rm -rf node_modules && npm install`
- Clear functions build: `rm -rf functions/lib`
- Rebuild: `npm run build`

## Next Steps

After Phase 1 is complete and tested, we'll implement:

1. **Phase 2**: Team member management (view, edit roles)
2. **Phase 3**: Create and manage feedback rounds
3. **Phase 4**: Feedback submission form
4. **Phase 5**: AI-powered consolidation with OpenAI
5. **Phase 6**: Share feedback with subjects
6. **Phase 7**: Polish, testing, and production deployment

## License

Private project - Not for distribution

## Support

For issues or questions, contact the development team.
