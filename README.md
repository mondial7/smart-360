# Smart 360 Feedback

A serverless web application for collecting, consolidating, and delivering anonymous 360-degree feedback using React + TypeScript + Firebase.

## Features

- 🔐 Google Sign-In authentication (admin + team member roles)
- 👥 Team member management
- 📝 Feedback rounds with deadline tracking
- 🤖 AI-powered feedback consolidation using OpenAI GPT-4o-mini
- 🔒 Complete anonymity for reviewers
- 📊 Beautiful dashboard with real-time updates

## Tech Stack

- **Frontend**: React 18 + TypeScript + Vite + Material-UI
- **Backend**: Firebase (Firestore, Auth, Cloud Functions)
- **AI**: OpenAI GPT-4o-mini for feedback consolidation
- **Deployment**: Firebase Hosting

## Prerequisites

- Node.js 18+ (Cloud Functions requirement)
- Firebase CLI (\`npm install -g firebase-tools\`)
- Firebase project (create at https://console.firebase.google.com)

## Getting Started

### 1. Install Dependencies

\`\`\`bash
# Install root dependencies
npm install

# Install Cloud Functions dependencies
cd functions
npm install
cd ..
\`\`\`

### 2. Configure Environment Variables

Create \`.env.local\` in the root directory:

\`\`\`env
VITE_FIREBASE_API_KEY=your_api_key
VITE_FIREBASE_AUTH_DOMAIN=your_project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your_project_id
VITE_FIREBASE_STORAGE_BUCKET=your_project.appspot.com
VITE_FIREBASE_MESSAGING_SENDER_ID=your_sender_id
VITE_FIREBASE_APP_ID=your_app_id
\`\`\`

### 3. Configure OpenAI API Key (for AI Consolidation)

\`\`\`bash
# For local development (emulators)
export OPENAI_API_KEY="sk-..."

# For production deployment
firebase functions:config:set openai.api_key="sk-..."
\`\`\`

### 4. Run Development Server

The \`npm dev\` command starts both the frontend (Vite) and Firebase emulators:

\`\`\`bash
npm dev
\`\`\`

This will start:
- **Frontend**: http://localhost:5173
- **Firebase Emulator UI**: http://localhost:4000
- **Firestore Emulator**: localhost:8080
- **Auth Emulator**: localhost:9099
- **Functions Emulator**: localhost:5001

### Individual Development Scripts

If you need to run services separately:

\`\`\`bash
# Frontend only
npm run dev:frontend

# Firebase emulators only
npm run dev:firebase

# Cloud Functions build watch mode (auto-rebuild on changes)
npm run dev:functions
\`\`\`

## Local Development

When you run \`npm dev\`, the Firebase emulators will:
- Import data from \`./firebase-export\` (if it exists)
- Export data on exit to preserve your local development data
- This allows you to maintain local test users and feedback rounds between sessions

**First Time Setup:**
1. Run \`npm dev\`
2. Open http://localhost:5173
3. Sign in with Google (creates your admin user automatically)
4. Create test team members and feedback rounds
5. When you stop the dev server (Ctrl+C), data is saved to \`firebase-export/\`
6. Next time you run \`npm dev\`, your data will be restored

## Building for Production

\`\`\`bash
# Build frontend
npm run build

# Build Cloud Functions
npm run build:functions

# Preview production build locally
npm run preview
\`\`\`

## Deployment

### Deploy Everything

\`\`\`bash
firebase deploy
\`\`\`

### Deploy Specific Services

\`\`\`bash
# Deploy only Firestore rules
firebase deploy --only firestore:rules

# Deploy only Cloud Functions
firebase deploy --only functions

# Deploy only hosting (frontend)
firebase deploy --only hosting
\`\`\`

## Available Scripts

- \`npm dev\` - Run frontend + Firebase emulators concurrently
- \`npm run dev:frontend\` - Run Vite development server only
- \`npm run dev:firebase\` - Run Firebase emulators only
- \`npm run dev:functions\` - Watch mode for Cloud Functions
- \`npm run build\` - Build frontend for production
- \`npm run build:functions\` - Build Cloud Functions
- \`npm run lint\` - Run ESLint
- \`npm run preview\` - Preview production build locally

## Firebase Emulators

The project uses Firebase Local Emulator Suite for development:

- **Auth Emulator** (port 9099): Test Google Sign-In locally
- **Firestore Emulator** (port 8080): Local database
- **Functions Emulator** (port 5001): Test Cloud Functions
- **Emulator UI** (port 4000): Visual interface for all emulators

## Security

- Firestore security rules enforce role-based access
- Only admins can read feedback submissions (ensures anonymity)
- Subjects can only read their own consolidated feedback
- Cloud Functions validate permissions server-side

## Phases Completed

✅ **Phase 1**: Project Foundation & Authentication  
✅ **Phase 2**: User Role Management  
✅ **Phase 3**: Feedback Rounds - Creation & Management  
✅ **Phase 4**: Feedback Submission  
✅ **Phase 5**: AI-Powered Feedback Consolidation

### Upcoming
- **Phase 6**: Feedback Delivery to Subject
- **Phase 7**: Polish & Production Readiness
