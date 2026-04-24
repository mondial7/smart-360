# Development Database Seeding

## Quick Start

Reset and reseed the database with comprehensive test data:

```bash
cd backend
./reseed-dev.sh
```

Or run directly:

```bash
cd backend
go run cmd/seed/main.go
```

## What Gets Created

The seed script creates a complete test environment:

### Users (6 total)
- **admin@example.com** - Admin (Emma Admin)
- **alice@example.com** - Member (Alice Johnson)
- **bob@example.com** - Member (Bob Smith)
- **carol@example.com** - Member (Carol Williams)
- **david@example.com** - Member (David Brown)
- **eve@example.com** - Member (Eve Martinez)

### Feedback Rounds (4 different statuses)

#### 1. DRAFT Round
- **Subject:** Alice
- **Status:** Draft (no reviewers assigned yet)
- **Use case:** Test adding reviewers and starting a round

#### 2. ACTIVE Round
- **Subject:** Bob
- **Status:** Active (in progress)
- **Deadline:** 5 days from now
- **Reviewers:** Alice, Carol, David (3 assigned)
- **Submissions:** 1 out of 3 submitted (Alice submitted)
- **Use case:** Test submitting feedback as reviewers

#### 3. CLOSED Round ⭐ **READY FOR AI CONSOLIDATION**
- **Subject:** Carol
- **Status:** Closed (deadline passed, all feedback collected)
- **Deadline:** 2 days ago
- **Reviewers:** Alice, Bob, David, Eve (4 assigned)
- **Submissions:** 4 out of 4 submitted ✅
- **Use case:** **Test AI-powered consolidation!**

#### 4. SHARED Round
- **Subject:** David
- **Status:** Shared (consolidation complete and shared)
- **Reviewers:** Alice, Bob, Carol (3 assigned)
- **Submissions:** 3 out of 3 submitted
- **Consolidation:** Already generated and shared
- **Use case:** See what a completed round looks like

## Testing AI Consolidation

### Prerequisites

1. Make sure you have a Gemini API key in your `.env` file:
   ```bash
   GEMINI_API_KEY=your-actual-api-key-here
   ```

2. Get a free Gemini API key at: https://makersuite.google.com/app/apikey

### Steps to Test

1. **Reseed the database**
   ```bash
   ./reseed-dev.sh
   ```
   The script will output the Round ID for Carol's closed round.

2. **Start the backend**
   ```bash
   go run main.go
   ```

3. **Login as admin**
   ```bash
   curl 'http://localhost:8080/api/auth/dev-login?email=admin@example.com'
   ```
   Copy the JWT token from the response.

4. **Trigger AI consolidation**
   ```bash
   # Replace {TOKEN} with your JWT token
   # Replace {ROUND_ID} with Carol's round ID from step 1

   curl -X POST http://localhost:8080/api/rounds/{ROUND_ID}/consolidate \
     -H "Authorization: Bearer {TOKEN}" \
     -H "Content-Type: application/json"
   ```

5. **View the consolidation**
   ```bash
   curl http://localhost:8080/api/consolidations/{ROUND_ID} \
     -H "Authorization: Bearer {TOKEN}"
   ```

### Expected AI Output

The AI should generate:
- **Executive Summary**: 2-3 sentence overview of Carol's feedback
- **Strengths**: Array of Carol's key strengths (communication, technical skills, mentoring, etc.)
- **Areas for Improvement**: Array of growth opportunities (public speaking, strategic thinking, etc.)
- **Actionable Insights**: Specific recommendations for Carol
- **Question Summaries**: Consolidated answers for each of the 4 feedback questions

### What the Realistic Feedback Contains

Carol's closed round has 4 detailed, realistic submissions covering:
- **Strengths**: Problem-solving, technical expertise, code quality, mentoring, attention to detail
- **Improvements**: Public speaking, strategic thinking, work-life balance, communication
- **Behaviors**: Reliable, helpful, thorough code reviews, high-quality deliverables
- **Advice**: Take on visible projects, improve documentation, delegate more, business impact focus

This realistic data helps test if the AI can:
- Extract common themes across multiple reviewers
- Distinguish between major and minor feedback points
- Generate actionable, specific recommendations
- Create a coherent executive summary
- Maintain professional, constructive tone

## Without Gemini API Key

If you don't have a Gemini API key, the consolidation will still work but will use a basic fallback that:
- Combines all feedback responses
- Creates simple summaries
- Groups feedback by question

This is useful for testing the UI and workflow without AI.

## Tips

- **Check the console output** when running `reseed-dev.sh` - it shows the Round ID you need for testing
- **Use the frontend** - it's much easier! Just login as admin@example.com and navigate to the closed round
- **API testing** - Use the backend's API endpoints directly to test without the frontend
- **Reset anytime** - Run `./reseed-dev.sh` whenever you want fresh data

## Troubleshooting

**MongoDB not starting?**
```bash
docker-compose up -d mongodb
docker ps  # Verify it's running
```

**Can't connect to MongoDB?**
Check your `.env` file has:
```
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=smart360
```

**Seed script fails?**
Make sure you're in the `backend` directory:
```bash
cd backend
./reseed-dev.sh
```

## File Structure

```
backend/
├── reseed-dev.sh           # Convenience script to reseed
├── cmd/seed/main.go        # Standalone seed program
├── database/
│   ├── seed.go            # Simple seed (used on app startup)
│   └── seed_dev.go        # Comprehensive dev seed
└── DEV-SEED-README.md     # This file
```
