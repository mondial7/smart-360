# AI Consolidation Feature - Setup Checklist (Archived)

> **Status: Archived — May 2026.** AI consolidation has been GA for some time and the setup steps
> now live in [`README.md`](../../README.md) and [`CLAUDE.md`](../../CLAUDE.md). Kept for
> historical reference only — round IDs and file paths in this document may be stale.

---

## ✅ What's Ready

### 1. Development Seeding Script ✅
- **Location**: `backend/cmd/seed/main.go` and `backend/database/seed_dev.go`
- **Status**: Tested and working!
- **What it creates**:
  - 6 users (1 admin, 5 members)
  - 4 feedback rounds in different statuses
  - **CLOSED round for Carol** with 4 realistic submissions ready for AI consolidation
  - Round ID: `69c95aeec387b43a272196bc` (will change each time you reseed)

### 2. Backend Code ✅
- **Consolidation endpoint**: `POST /api/rounds/:id/consolidate` ✅
- **Gemini integration**: Fully implemented in `backend/handlers/consolidation.go` ✅
- **Dependencies**: `github.com/google/generative-ai-go` already in go.mod ✅
- **Fallback**: Works without Gemini API key (uses basic text combination) ✅

### 3. Database ✅
- **MongoDB**: Running at localhost:27017 ✅
- **Collections**: All required collections created ✅
- **Test data**: Rich, realistic feedback data seeded ✅

## ⚠️ What's Missing (Required for AI)

### 1. Environment Configuration ❌
You need to create a `.env` file in the `backend` directory.

**Quick Setup:**
```bash
cd backend
cp .env.example .env
```

**Then edit `.env` and add your Gemini API key:**
```bash
# MongoDB Configuration (should work as-is)
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=smart360

# OAuth Configuration (needed for Google login, but not for dev-login)
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback

# Application Configuration
FRONTEND_URL=http://localhost:5173
JWT_SECRET=your-random-secret-key-change-in-production

# AI Configuration (REQUIRED FOR AI CONSOLIDATION)
GEMINI_API_KEY=your-actual-gemini-api-key-here
```

### 2. Gemini API Key ⚠️
**Where to get it:**
1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Create API Key"
4. Copy the key and paste it in your `.env` file

**Note:** The free tier of Gemini is very generous and should be more than enough for development/testing.

## 🚀 Quick Start Guide

### Step 1: Setup Environment
```bash
cd backend
cp .env.example .env
# Edit .env and add your GEMINI_API_KEY
```

### Step 2: Seed the Database
```bash
./reseed-dev.sh
# Or: go run cmd/seed/main.go
```

**Copy the Round ID from the output** (it will look like: `69c95aeec387b43a272196bc`)

### Step 3: Start the Backend
```bash
go run main.go
```

### Step 4: Test AI Consolidation

**Option A: Using the Frontend (Easiest)**
1. Start the frontend: `cd ../frontend && npm run dev`
2. Open http://localhost:5173
3. Login with dev mode: Use dev-login or navigate to `/api/auth/dev-login?email=admin@example.com`
4. Go to the closed round for Carol
5. Click "Consolidate Feedback"
6. Watch the AI magic happen! ✨

**Option B: Using curl (Direct API testing)**
```bash
# 1. Login as admin
TOKEN=$(curl -s 'http://localhost:8080/api/auth/dev-login?email=admin@example.com' | jq -r '.token')

# 2. Trigger consolidation (replace ROUND_ID with the one from seeding output)
curl -X POST http://localhost:8080/api/rounds/69c95aeec387b43a272196bc/consolidate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# 3. View the result
curl http://localhost:8080/api/consolidations/69c95aeec387b43a272196bc \
  -H "Authorization: Bearer $TOKEN" | jq
```

## 📝 What the AI Should Generate

When consolidation works correctly, you should see:

### Executive Summary
A 2-3 sentence overview synthesizing all the feedback for Carol.

### Strengths (Array)
- Problem-solving skills
- Technical expertise
- Code quality and attention to detail
- Mentoring abilities
- Thorough code reviews
- etc.

### Areas for Improvement (Array)
- Public speaking and presentation skills
- Strategic thinking vs. technical details
- Work-life balance
- Proactive knowledge sharing
- Communication about progress

### Actionable Insights (Array)
- Specific, concrete recommendations Carol can implement
- Example: "Consider taking on more visible projects to showcase your talents"
- Example: "Schedule regular sync-ups with stakeholders"

### Question Summaries (Object)
Consolidated answers for each of the 4 feedback questions:
- `a`: Strengths summary
- `b`: Improvements summary
- `c`: Behaviors summary
- `d`: Growth advice summary

## 🧪 Testing Without Gemini API Key

If you don't have a Gemini API key yet, the feature will still work with a basic fallback:
- It combines all feedback text
- Creates simple summaries
- Groups responses by question
- No AI intelligence, but tests the full workflow

This is useful for:
- Testing the UI/UX
- Verifying the data flow
- Making sure all endpoints work
- Development when offline

## 🐛 Troubleshooting

### "No .env file found"
- This is OK if MongoDB environment variables are already set
- But you MUST create `.env` with `GEMINI_API_KEY` for AI features

### "Failed to generate AI consolidation"
- Check your Gemini API key is correct
- Verify you have internet connection
- Check console logs for specific error
- Make sure you haven't exceeded API quota (unlikely on free tier)

### "No feedback submissions found to consolidate"
- Make sure you're using the CLOSED round ID from the seeding output
- Verify the database was seeded correctly: `go run cmd/seed/main.go`
- Check there are submissions: `GET /api/submissions/round/{roundId}`

### MongoDB connection issues
- Verify MongoDB is running: `docker ps | grep mongo`
- Start if needed: `docker-compose up -d mongodb`
- Check connection string in `.env` matches your setup

## 📚 Additional Resources

- **Development seed README**: `backend/DEV-SEED-README.md`
- **Product overview**: `PRODUCT-OVERVIEW.md`
- **Consolidation handler code**: `backend/handlers/consolidation.go`
- **Gemini API docs**: https://ai.google.dev/docs

## ✨ Summary

**What works right now (without any changes):**
- ✅ Database seeding with realistic test data
- ✅ All API endpoints
- ✅ Complete feedback submission workflow
- ✅ Basic consolidation (without AI)

**What needs 1 minute of setup for AI:**
- ⚠️ Create `.env` file
- ⚠️ Add Gemini API key

That's it! Once you have the API key, you're ready to test the full AI-powered consolidation feature.
