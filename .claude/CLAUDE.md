# Smart 360 Feedback - Development Guide

This guide is designed for developers (including AI assistants like Claude) working on the Smart 360 Feedback codebase. It provides architectural context, development workflows, and common tasks.

## Project Overview

Smart 360 Feedback is an AI-powered anonymous peer feedback platform built with:
- **Backend:** Go 1.25.1 + Gin framework + MongoDB
- **Frontend:** Vue.js 3 + TypeScript + Vite
- **AI:** Google Gemini API for feedback consolidation
- **Auth:** Google OAuth 2.0 + JWT

The application enables team members to submit anonymous feedback which is then synthesized into actionable insights using AI.

## Architecture

### Backend (`/backend/`)

**Framework:** Gin (Go web framework)
**Database:** MongoDB 8.0 with mongo-driver
**Authentication:** Google OAuth 2.0 + JWT
**AI Integration:** Google Gemini API (genai package)
**Port:** 8080

#### Directory Structure

```
backend/
├── cmd/              # CLI commands (seeding utilities)
├── database/         # DB initialization, seeding, indexes
├── handlers/         # HTTP request handlers (route logic)
├── middleware/       # Auth middleware, role checks, CORS
├── models/           # Data structures and MongoDB models
├── scripts/          # MongoDB init scripts
├── main.go           # Application entry point
├── go.mod            # Go dependencies
└── .env              # Environment variables (not in git)
```

#### Key Files

- **main.go** - Server entry point, routes, middleware setup
- **database/db.go** - MongoDB connection initialization
- **database/seed_dev.go** - Test data seeding for development
- **handlers/auth.go** - OAuth and JWT logic
- **handlers/consolidation.go** - AI-powered feedback consolidation
- **middleware/auth.go** - JWT verification middleware

### Frontend (`/frontend/`)

**Framework:** Vue.js 3 (Composition API)
**Language:** TypeScript
**Build Tool:** Vite
**State Management:** Pinia
**Routing:** Vue Router
**HTTP Client:** Axios
**Styling:** SASS
**Port:** 5173 (dev), 80 (production via nginx)

#### Directory Structure

```
frontend/
├── src/
│   ├── api/          # Axios client configuration
│   ├── assets/       # Static assets (SCSS, images)
│   ├── components/   # Reusable Vue components
│   ├── router/       # Vue Router configuration
│   ├── stores/       # Pinia state stores
│   ├── types/        # TypeScript type definitions
│   ├── views/        # Page components
│   ├── App.vue       # Root component
│   └── main.ts       # App entry point
├── package.json      # npm dependencies & scripts
└── vite.config.ts    # Vite configuration (port, proxy)
```

#### Key Files

- **src/main.ts** - App initialization
- **src/App.vue** - Root component with navbar/router
- **src/router/index.ts** - Route definitions
- **src/api/client.ts** - Axios configuration with JWT interceptor
- **src/stores/auth.ts** - Auth state management (user, token, login/logout)

### Database Schema

**Database:** smart360
**Collections:**

1. **users** - User profiles
   - Fields: email, name, role, photo_url, created_at
   - Roles: `admin`, `team_admin`, `member`

2. **teams** - Team definitions
   - Fields: name, description, team_admin_id, created_at

3. **team_members** - Team membership junction table
   - Fields: team_id, user_id

4. **feedback_rounds** - Feedback cycles
   - Fields: subject_id, status, deadline, created_by, created_at
   - Statuses: `draft`, `active`, `closed`, `shared`

5. **round_reviewers** - Reviewer assignments
   - Fields: round_id, reviewer_id, submitted

6. **submissions** - Anonymous feedback responses
   - Fields: round_id, responses (JSON with keys: a, b, c, d), submitted_at

7. **consolidations** - AI-generated insights
   - Fields: round_id, executive_summary, strengths, areas_for_improvement, actionable_insights, question_summaries

8. **audit_logs** - Action tracking
   - Fields: user_id, action, entity_type, entity_id, timestamp

**Indexes:** Defined in `backend/database/indexes.go` for performance optimization.

## Development Workflow

### Initial Setup

#### 1. Start MongoDB

```bash
cd backend
docker-compose up -d mongodb
```

This starts MongoDB on port 27017 with credentials `admin:password123`.

#### 2. Configure Environment

```bash
cp backend/.env.example backend/.env
# Edit backend/.env with your Google OAuth and Gemini API credentials
```

Required variables:
- `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` (from Google Cloud Console)
- `GEMINI_API_KEY` (from Google AI Studio)
- `JWT_SECRET` (generate with `openssl rand -base64 32`)

#### 3. Seed Development Data

```bash
cd backend
./reseed-dev.sh
```

This creates:
- 6 test users (admin@example.com is admin)
- 4 feedback rounds with realistic data
- Pre-submitted feedback for testing

#### 4. Start Backend

```bash
cd backend
go run main.go
# Server starts on :8080
```

#### 5. Start Frontend

```bash
cd frontend
npm install
npm run dev
# Dev server starts on :5173
```

#### 6. Access Application

- Navigate to http://localhost:5173
- Use dev-login: http://localhost:8080/api/auth/dev-login?email=admin@example.com
- Or use Google OAuth if configured

### Development with Docker

To test the full Docker setup:

```bash
# From project root
docker-compose up --build

# Access at http://localhost (port 80)
```

## Code Style Guidelines

### Go (Backend)

**Conventions:**
- Follow standard Go conventions (use `gofmt`, `golint`)
- Use meaningful variable names (avoid single letters except in loops)
- Always check errors, return early
- Comment exported functions
- Keep handlers thin, move business logic to separate functions

**Error Handling:**
```go
// Good
result, err := database.GetUser(id)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
    return
}

// Bad - ignoring errors
result, _ := database.GetUser(id)
```

**Handler Pattern:**
```go
func HandlerName(c *gin.Context) {
    // 1. Parse/validate input
    // 2. Call business logic
    // 3. Return response
    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

### TypeScript (Frontend)

**Conventions:**
- Use Composition API with `<script setup lang="ts">`
- Type everything - avoid `any`
- PascalCase for components, camelCase for functions/variables
- Use SASS with scoped styles: `<style scoped lang="scss">`
- API calls via `apiClient` from `src/api/client.ts`

**Component Structure:**
```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { User } from '@/types/user'

const users = ref<User[]>([])

const fetchUsers = async () => {
  // Logic here
}
</script>

<template>
  <!-- Template here -->
</template>

<style scoped lang="scss">
/* Scoped styles here */
</style>
```

## Testing Approach

### Backend Testing

- **Manual testing** via Postman/curl
- **Dev login endpoint:** `/api/auth/dev-login?email=<email>` (bypasses OAuth)
- **Test data:** Use `reseed-dev.sh` for consistent state
- **Debug endpoints:**
  - `/api/debug/submissions` - View all submissions
  - `/api/debug/reviewers` - View all reviewer assignments (admin only)

### Frontend Testing

- **Manual testing** in browser
- **Vue DevTools** for component inspection
- **Network tab** for API debugging
- **Console logs** for state debugging

## Common Tasks

### Add a New API Endpoint

1. **Define handler** in `backend/handlers/<domain>.go`:

```go
func GetFeedbackStats(c *gin.Context) {
    // Handler logic
    c.JSON(http.StatusOK, gin.H{"stats": stats})
}
```

2. **Add route** in `backend/main.go`:

```go
authorized.GET("/feedback/stats", handlers.GetFeedbackStats)
// Or with middleware:
authorized.GET("/feedback/stats", middleware.AdminOnly(), handlers.GetFeedbackStats)
```

3. **Test** with curl:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/feedback/stats
```

4. **Add frontend API call** in appropriate view or store.

### Add a New Vue Page

1. **Create view** in `frontend/src/views/<Name>View.vue`:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'

onMounted(() => {
  // Fetch data
})
</script>

<template>
  <div class="page-container">
    <h1>Page Title</h1>
  </div>
</template>

<style scoped lang="scss">
.page-container {
  padding: 2rem;
}
</style>
```

2. **Add route** in `frontend/src/router/index.ts`:

```typescript
{
  path: '/new-page',
  name: 'NewPage',
  component: () => import('@/views/NewPageView.vue'),
  meta: { requiresAuth: true }
}
```

3. **Add navigation link** in `frontend/src/components/NavBar.vue` if needed.

### Modify Database Schema

1. **Update model** in `backend/models/<model>.go`:

```go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Email     string             `bson:"email" json:"email"`
    NewField  string             `bson:"new_field" json:"newField"` // Add this
    // ...
}
```

2. **Update MongoDB init script** `backend/scripts/mongo-init.js` (if schema validation needed)

3. **Update indexes** in `backend/database/indexes.go` if querying new field

4. **Update seed data** in `backend/database/seed_dev.go`

5. **Drop and recreate database:**

```bash
docker-compose down -v  # Removes volumes
docker-compose up -d
cd backend && ./reseed-dev.sh
```

### Run Database Seeding

```bash
cd backend
./reseed-dev.sh
```

This script:
- Drops all collections
- Recreates them with fresh seed data
- Creates 6 users, 4 rounds, realistic feedback

## Environment Variables

### Backend (.env in /backend/)

| Variable | Description | Example |
|----------|-------------|---------|
| `MONGODB_URI` | MongoDB connection string | `mongodb://admin:password123@localhost:27017` |
| `MONGODB_DB` | Database name | `smart360` |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | `xxx.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth secret | `GOCSPX-xxx` |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL | `http://localhost:8080/api/auth/callback` |
| `FRONTEND_URL` | Frontend URL for CORS | `http://localhost:5173` |
| `JWT_SECRET` | Secret for JWT signing | `random-secret-key` |
| `GEMINI_API_KEY` | Google Gemini API key | `AIzaSyXXX` |

### Frontend

No .env needed - uses Vite proxy:
- API calls to `/api` are proxied to `http://localhost:8080`
- Configured in `vite.config.ts`

## Key Features

### Authentication Flow

1. User clicks "Login with Google"
2. Frontend redirects to `/api/auth/google`
3. Backend redirects to Google OAuth
4. Google redirects to `/api/auth/callback`
5. Backend creates/updates user, generates JWT
6. Redirects to frontend with token in query param
7. Frontend stores token in localStorage
8. All API requests include `Authorization: Bearer <token>` header

### Feedback Round Lifecycle

1. **Draft**: Admin creates round, adds reviewers
2. **Active**: Admin starts round, reviewers can submit feedback
3. **Closed**: Round auto-closes at deadline (or manually closed)
4. **Shared**: Admin generates consolidation, reviews, and shares with subject

### AI Consolidation Process

1. Fetch all submissions for round
2. Parse responses (JSON with keys: a, b, c, d)
3. Format as prompt for Gemini
4. Call Gemini API with structured output request
5. Parse JSON response:
   - `executive_summary` (string)
   - `strengths` (array of strings)
   - `areas_for_improvement` (array of objects with area + details)
   - `actionable_insights` (array of objects with insight + rationale)
   - `question_summaries` (object with a, b, c, d keys)
6. Save to consolidations collection
7. Update round status

## Important Files

### Configuration

- `/backend/.env` - Backend environment variables (not in git)
- `/backend/.env.example` - Template for .env
- `/backend/docker-compose.yml` - MongoDB setup (dev)
- `/frontend/vite.config.ts` - Vite config with API proxy
- `/docker-compose.yml` - Full stack Docker config (root)

### Key Backend Files

- `/backend/main.go` - Server entry point, routes, CORS
- `/backend/database/db.go` - MongoDB connection
- `/backend/database/seed_dev.go` - Test data seeding
- `/backend/handlers/auth.go` - OAuth and JWT logic
- `/backend/handlers/consolidation.go` - AI consolidation
- `/backend/middleware/auth.go` - JWT verification
- `/backend/middleware/admin.go` - Admin role check

### Key Frontend Files

- `/frontend/src/main.ts` - App entry point
- `/frontend/src/App.vue` - Root component
- `/frontend/src/router/index.ts` - Route definitions
- `/frontend/src/api/client.ts` - Axios configuration
- `/frontend/src/stores/auth.ts` - Auth state management
- `/frontend/src/views/DashboardView.vue` - Main dashboard
- `/frontend/src/views/RoundsManagement.vue` - Admin round management

## Development Tips

1. **CORS issues**: Backend CORS is configurable via `FRONTEND_URL` env var
2. **MongoDB auth**: Default credentials are `admin:password123`
3. **Dev login**: Use `/api/auth/dev-login?email=<email>` to bypass OAuth
4. **Reseed often**: Run `./reseed-dev.sh` to reset to known good state
5. **Check logs**: Backend logs to stdout, frontend shows in browser console
6. **Gemini fallback**: Without API key, consolidation uses basic text combination
7. **First user is admin**: First person to login gets admin role automatically

## Debugging

### Backend Issues

```bash
# View logs
docker-compose logs backend -f

# Check MongoDB connection
docker exec -it smart360-mongodb mongosh -u admin -p password123

# Test endpoint directly
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/me
```

### Frontend Issues

```bash
# Clear Vite cache
rm -rf frontend/.vite frontend/node_modules/.vite

# Rebuild
cd frontend && npm run build

# Check API proxy
# In browser DevTools Network tab, verify /api calls go to localhost:8080
```

### Common Errors

**"MongoDB connection failed"**
- Ensure MongoDB is running: `docker-compose ps mongodb`
- Check credentials in .env match docker-compose.yml

**"CORS policy error"**
- Update `FRONTEND_URL` in backend/.env to match frontend URL
- For Docker: `FRONTEND_URL=http://localhost`
- For dev: `FRONTEND_URL=http://localhost:5173`

**"Invalid token"**
- Token might be expired (24h default)
- Re-login to get new token
- Check JWT_SECRET matches between sessions

**"Consolidation failed"**
- Check GEMINI_API_KEY is set
- Verify API quota isn't exceeded
- Check backend logs for detailed error

## Production Considerations

1. **Change all default passwords** (MongoDB, JWT secret)
2. **Use proper Google OAuth credentials** (not localhost)
3. **Set secure CORS origins**
4. **Enable HTTPS** (use reverse proxy like nginx or Caddy)
5. **Set appropriate MongoDB user permissions** (not root)
6. **Monitor Gemini API usage** and costs
7. **Implement rate limiting** for API endpoints
8. **Add proper logging** (structured logging, not just stdout)
9. **Set up database backups** (MongoDB dump strategy)
10. **Use environment-specific .env files** (dev/staging/prod)

## Resources

- **Go Gin Documentation:** https://gin-gonic.com/docs/
- **Vue 3 Documentation:** https://vuejs.org/guide/introduction.html
- **MongoDB Go Driver:** https://www.mongodb.com/docs/drivers/go/current/
- **Google Gemini API:** https://ai.google.dev/docs
- **Vite Documentation:** https://vitejs.dev/guide/
- **Pinia Documentation:** https://pinia.vuejs.org/

## Git Commit Guidelines

- Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- Write clear, descriptive commit messages
- Keep commits focused and atomic
- Do NOT include Claude as commit co-author (per user preferences)

## Notes for AI Assistants

- When modifying backend code, maintain existing error handling patterns
- When adding frontend features, use Composition API with TypeScript
- Always update seed data if changing database schema
- Test changes with dev-login before requiring OAuth setup
- Follow existing naming conventions (handlers vs controllers, etc.)
- Backend uses MongoDB ObjectIDs - always convert from hex strings
- Frontend stores JWT in localStorage - consider this for auth flows
- Gemini API calls are in consolidation.go - reference for prompt engineering
- The project prioritizes simplicity over premature optimization
