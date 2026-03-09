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
✅ **Phase 6**: Feedback Delivery to Subject
✅ **Phase 7**: Polish & Production Readiness

## Production Deployment Checklist

### 1. Firebase Project Setup

Ensure your Firebase project is on the **Blaze (Pay-as-you-go) plan** for Cloud Functions.

### 2. Configure Environment Variables

Set production environment variables:

\`\`\`bash
# Set OpenAI API key for Cloud Functions
firebase functions:config:set openai.api_key="sk-your-production-key"

# Set app URL for email links
firebase functions:config:set app.url="https://your-domain.com"

# Optional: Configure email service (SendGrid, AWS SES, etc.)
firebase functions:config:set sendgrid.api_key="your-sendgrid-key"
\`\`\`

### 3. Deploy Firestore Indexes

Deploy indexes before deploying functions to avoid query errors:

\`\`\`bash
firebase deploy --only firestore:indexes
\`\`\`

### 4. Deploy Security Rules

\`\`\`bash
firebase deploy --only firestore:rules
\`\`\`

### 5. Build and Deploy Cloud Functions

\`\`\`bash
npm run build:functions
firebase deploy --only functions
\`\`\`

### 6. Build and Deploy Frontend

\`\`\`bash
npm run build
firebase deploy --only hosting
\`\`\`

### 7. Deploy Everything

Or deploy everything at once:

\`\`\`bash
npm run build
npm run build:functions
firebase deploy
\`\`\`

### 8. Post-Deployment Configuration

#### Enable Email Notifications (Optional)

To enable email notifications, integrate with an email service:

**Option A: SendGrid**
\`\`\`bash
cd functions
npm install @sendgrid/mail
firebase functions:config:set sendgrid.api_key="your-key"
\`\`\`

Then uncomment the SendGrid code in \`functions/src/services/emailService.ts\`

**Option B: Gmail SMTP**
Use nodemailer with Gmail App Password (not recommended for production)

**Option C: AWS SES**
Use AWS SES for reliable email delivery

#### Configure Custom Domain (Optional)

1. Go to Firebase Console → Hosting
2. Add custom domain
3. Follow DNS configuration instructions
4. Update \`app.url\` config: \`firebase functions:config:set app.url="https://your-domain.com"\`

#### Set Up Monitoring

1. Enable Firebase Performance Monitoring
2. Set up Cloud Function alerts in Firebase Console
3. Configure budget alerts in Google Cloud Console
4. Monitor Firestore usage and quotas

### 9. Testing in Production

1. Sign in with your Google account (first user becomes admin)
2. Test complete feedback flow:
   - Create team members
   - Create feedback round
   - Submit feedback as reviewer
   - Consolidate with AI
   - Share with subject
3. Verify email notifications (if enabled)
4. Check Cloud Function logs for errors

## Email Notifications

The app includes email notification triggers for:

- **Feedback Request**: When reviewers are assigned to a feedback round
- **Deadline Reminders**: Automated reminders 3 days and 1 day before deadline (runs daily at 9 AM)
- **Feedback Shared**: When admin shares consolidated feedback with subject

Email notifications are logged by default. To enable actual sending:
1. Configure an email service (SendGrid, AWS SES, etc.)
2. Update \`functions/src/services/emailService.ts\` with your provider
3. Redeploy Cloud Functions

## Performance & Scalability

- **Firestore Indexes**: All necessary composite indexes are defined in \`firestore.indexes.json\`
- **Cloud Functions**: Use Node.js 18, optimized for cold starts
- **Frontend**: Vite build with code splitting and lazy loading ready
- **Caching**: Client-side caching via Firebase SDK
- **Real-time Updates**: Firestore listeners for live data sync

## Cost Optimization

- **Firestore**: ~$0.06 per 100K reads, $0.18 per 100K writes
- **Cloud Functions**: Free tier: 2M invocations/month, 400K GB-seconds/month
- **OpenAI**: GPT-4o-mini is ~$0.15 per 1M input tokens (~$0.005 per consolidation)
- **Hosting**: Free for most use cases (10 GB transfer/month)

**Estimated monthly cost for small team (20 users, 10 rounds/month):**
- Firestore: ~$1-2
- Cloud Functions: Free tier
- OpenAI: ~$0.05
- **Total: ~$1-2/month**

## Troubleshooting

### Cloud Functions not triggering

\`\`\`bash
# Check function logs
firebase functions:log

# Verify function deployment
firebase functions:list
\`\`\`

### Firestore permission denied

1. Verify you're signed in
2. Check Firestore rules are deployed: \`firebase deploy --only firestore:rules\`
3. Verify user role in Firestore console

### Email notifications not sending

1. Check Cloud Function logs: \`firebase functions:log --only sendDeadlineReminders\`
2. Verify email service configuration
3. Check \`functions:config\` settings: \`firebase functions:config:get\`

### OpenAI API errors

1. Verify API key: \`firebase functions:config:get openai\`
2. Check API key has sufficient credits
3. Review function logs for specific error messages
