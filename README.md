# Smart 360 Feedback

> AI-powered anonymous peer feedback platform for professional development

## What is Smart 360 Feedback?

Smart 360 Feedback is a web application that enables anonymous peer feedback within organizations, using Google Gemini AI to transform raw feedback into actionable insights. Team members can submit honest, constructive feedback anonymously, and administrators can generate AI-powered consolidated reports that highlight strengths, areas for improvement, and actionable recommendations.

### Key Features

- **Anonymous Feedback Submission** - 4-question structured format ensures comprehensive feedback
- **AI-Powered Consolidation** - Google Gemini analyzes and synthesizes feedback into actionable insights
- **PDF Export** - Download a branded, print-ready PDF of any consolidated feedback shared with you
- **Personal Analytics** - Radar chart of your latest round and a strengths/improvements/insights trend across rounds, on the user dashboard
- **Google OAuth Authentication** - Secure, familiar login experience
- **Role-Based Access Control** - Admin, Team Admin, and Member roles with appropriate permissions
- **Automated Round Management** - Create, schedule, and manage feedback rounds with ease
- **Real-Time Dashboards** - Track feedback status, pending submissions, and received feedback
- **Team Management** - Organize users into teams with dedicated team administrators
- **Audit Logging** - Track every status transition with the actor, timestamp, and round context

## Technology Stack

- **Backend:** Go 1.25.1 + Gin framework + MongoDB
- **Frontend:** Vue.js 3 + TypeScript + Vite
- **Database:** MongoDB 8.0
- **AI:** Google Gemini API
- **Deployment:** Docker + Docker Compose

## Quick Start

### Prerequisites

Before you begin, ensure you have:

- **Docker** and **Docker Compose** installed ([Get Docker](https://docs.docker.com/get-docker/))
- **Google OAuth credentials** ([Setup Guide](#google-oauth-setup))
- **Gemini API key** ([Get API Key](#gemini-api-setup))

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/yourusername/smart-360-in-go.git
cd smart-360-in-go
```

2. **Configure environment variables**

```bash
cp .env.example .env
```

Edit `.env` with your favorite editor and fill in the required values:

```bash
nano .env
# or
vim .env
# or
code .env
```

3. **Start the application**

```bash
docker-compose up -d
```

This will start all three services (MongoDB, Backend, Frontend) and initialize the database.

4. **Access the application**

- **Frontend:** http://localhost
- **Backend API:** http://localhost:8080 (internal only, proxied by frontend)

### First-Time Setup

1. Navigate to http://localhost and click "Login with Google"
2. The first user to log in will automatically become an **administrator**
3. Admins can promote other users to admin or team admin roles via the Users page
4. Start creating feedback rounds from the dashboard!

## Configuration

All configuration is managed via the `.env` file in the root directory. Below is a complete reference:

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| `MONGO_ROOT_USER` | Yes | MongoDB admin username | `admin` |
| `MONGO_ROOT_PASSWORD` | Yes | MongoDB admin password | `password123` |
| `GOOGLE_CLIENT_ID` | Yes | Google OAuth client ID | - |
| `GOOGLE_CLIENT_SECRET` | Yes | Google OAuth client secret | - |
| `GOOGLE_REDIRECT_URL` | No | OAuth callback URL | `http://localhost:8080/api/auth/callback` |
| `FRONTEND_URL` | No | Frontend URL for CORS | `http://localhost` |
| `FRONTEND_PORT` | No | Port to expose frontend | `80` |
| `JWT_SECRET` | Yes | Secret for JWT token signing | - |
| `GEMINI_API_KEY` | Yes | Google Gemini API key | - |

### Security Notes

- **Change default passwords** - Update `MONGO_ROOT_PASSWORD` and `JWT_SECRET` before deploying
- **Generate strong JWT secret** - Use `openssl rand -base64 32` to generate a random secret
- **Protect your .env file** - Never commit `.env` to version control

## Google OAuth Setup

To enable Google authentication, you'll need to create OAuth credentials:

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project or select an existing one
3. Click **"Create Credentials"** → **"OAuth client ID"**
4. Choose **"Web application"**
5. Add **Authorized redirect URI**: `http://localhost:8080/api/auth/callback`
6. For production, add your domain: `https://yourdomain.com/api/auth/callback`
7. Copy the **Client ID** and **Client Secret** to your `.env` file

For detailed setup instructions, see the [Google OAuth documentation](https://developers.google.com/identity/protocols/oauth2).

## Gemini API Setup

The AI consolidation feature requires a Gemini API key:

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click **"Create API Key"**
4. Copy the key to your `.env` file as `GEMINI_API_KEY`

**Note:** The Gemini API has usage limits. Monitor your usage in the [Google Cloud Console](https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/quotas).

## Docker Commands

### Basic Operations

```bash
# Start all services
docker-compose up -d

# View logs (all services)
docker-compose logs -f

# View logs (specific service)
docker-compose logs -f backend

# Stop all services
docker-compose down

# Stop and remove all data (WARNING: deletes database!)
docker-compose down -v
```

### Rebuilding After Code Changes

```bash
# Rebuild and restart
docker-compose up -d --build

# Rebuild specific service
docker-compose up -d --build backend
```

### Service Management

```bash
# View service status
docker-compose ps

# Restart a specific service
docker-compose restart backend

# View resource usage
docker stats
```

## Development

For development setup, building from source, and contributing guidelines, see:

- **Project Guide:** [CLAUDE.md](CLAUDE.md) — architecture, conventions, and key tasks
- **Contributing:** We welcome contributions! Please open an issue to discuss before submitting PRs.

## Troubleshooting

### Frontend shows connection error

**Symptoms:** Frontend loads but shows "Unable to connect to server" or API errors

**Solutions:**
1. Check backend is running: `docker-compose ps`
2. View backend logs: `docker-compose logs backend`
3. Ensure MongoDB is healthy: `docker-compose ps mongodb`
4. Verify environment variables in `.env`

### OAuth login fails

**Symptoms:** Redirected to Google but error after authorization

**Solutions:**
1. Verify `GOOGLE_REDIRECT_URL` in `.env` matches OAuth credentials
2. Check `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are correct
3. Ensure frontend URL is added to "Authorized JavaScript origins" in Google Console
4. Check backend logs for detailed error: `docker-compose logs backend`

### AI consolidation fails

**Symptoms:** "Failed to generate consolidation" error

**Solutions:**
1. Verify `GEMINI_API_KEY` is set correctly in `.env`
2. Check Gemini API quota: [Google Cloud Console](https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/quotas)
3. View backend logs for detailed error: `docker-compose logs backend`
4. Ensure internet connectivity from Docker containers

### Port 80 already in use

**Symptoms:** `docker-compose up` fails with "port 80 is already in use"

**Solutions:**
1. Change `FRONTEND_PORT` in `.env` to another port (e.g., `3000`)
2. Access frontend at `http://localhost:3000`
3. Or, stop the service using port 80 (e.g., Apache, nginx)

### Database connection fails

**Symptoms:** Backend logs show MongoDB connection errors

**Solutions:**
1. Ensure MongoDB container is running: `docker-compose ps mongodb`
2. Check MongoDB logs: `docker-compose logs mongodb`
3. Verify MongoDB credentials in `.env`
4. Wait for MongoDB health check to pass (can take 20-30 seconds on first start)

### Services start in wrong order

**Symptoms:** Backend fails because MongoDB isn't ready

**Solutions:**
1. Docker Compose uses health checks to ensure proper startup order
2. If issues persist, restart: `docker-compose restart backend`
3. Check health check status: `docker-compose ps`

## Architecture

```
┌─────────────────────────────────────────┐
│          Frontend (nginx:80)            │
│      Vue.js 3 + TypeScript + Vite       │
│         Serves static files             │
└──────────────────┬──────────────────────┘
                   │
          /api requests proxied to backend
                   │
┌──────────────────▼──────────────────────┐
│         Backend (:8080)                 │
│       Go + Gin Framework                │
│   REST API + OAuth + JWT + Gemini AI    │
└──────────────────┬──────────────────────┘
                   │
            mongodb://mongodb:27017
                   │
┌──────────────────▼──────────────────────┐
│         MongoDB (:27017)                │
│        Document Database                │
│   Collections: users, rounds,           │
│   submissions, consolidations           │
└─────────────────────────────────────────┘
```

## User Roles

| Role | Capabilities |
|------|-------------|
| **Admin** | Full access: Create rounds, assign reviewers, consolidate feedback, manage users, manage teams |
| **Team Admin** | Manage team members, create rounds for team, view team feedback |
| **Member** | Submit feedback, view received feedback, participate in rounds |

## How It Works

### 1. Create a Feedback Round

Admins create a feedback round for a subject (the person receiving feedback):
- Select the subject
- Choose reviewers (peers who will submit feedback)
- Set a deadline
- Start the round

### 2. Submit Feedback

Reviewers receive 4 questions about the subject:
- What does this person do well?
- What could they improve?
- What specific actions would help their growth?
- Additional comments

Submissions are **anonymous** - the subject won't know who wrote what.

### 3. AI Consolidation

Once all feedback is submitted (or the deadline passes), admins trigger AI consolidation:
- Google Gemini analyzes all responses
- Identifies common themes and patterns
- Generates executive summary
- Lists key strengths
- Highlights areas for improvement
- Provides actionable insights

### 4. Share Feedback

Admins review the consolidation, optionally add notes, and share it with the subject. The subject can then view their comprehensive feedback report in the dashboard.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation:** [PRD.md](PRD.md) | [PRODUCT-OVERVIEW.md](PRODUCT-OVERVIEW.md)
- **Issues:** [GitHub Issues](https://github.com/yourusername/smart-360-in-go/issues)
- **Project Guide:** [CLAUDE.md](CLAUDE.md)

## Roadmap

See [NEXT_STEPS.md](NEXT_STEPS.md) for planned features and improvements, including:

- Admin-level analytics dashboard
- 360 comparison across rounds
- Anonymous comments on shared feedback
- Slack / Teams integrations
- And more

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

Please read [CLAUDE.md](CLAUDE.md) for development setup and coding guidelines.

---

**Built with ❤️ using Go, Vue.js, and Google Gemini AI**
