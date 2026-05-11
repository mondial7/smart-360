# Smart 360 Feedback - Product Overview

## What It Is
**Smart 360 Feedback** is a web application that enables anonymous peer feedback within organizations, using AI to transform raw feedback into actionable insights for professional development.

---

## Key Features

### For Administrators
- **Feedback Round Management** - Create and coordinate feedback cycles with customizable deadlines
- **Team Roster** - Centralized user management with role-based access control (Admin / Team Admin / Member)
- **Team-Scoped Rounds** - Team admins manage feedback within their team
- **AI-Powered Consolidation** - Google Gemini synthesizes feedback into structured insights automatically
- **Feedback Distribution** - Control when and how consolidated feedback is shared with team members
- **Real-Time Tracking** - Monitor submission status and engagement across active rounds
- **Audit Logging** - Track every status transition with the actor, timestamp, and round context

### For Team Members
- **Anonymous Feedback Submission** - Provide honest, structured responses through 4-question format
- **Personal Dashboard** - Track pending reviews and received feedback in one place
- **Personal Analytics** - Radar chart of your latest round and a strengths/improvements/insights trend over time
- **PDF Export** - Download a branded PDF of any consolidated feedback shared with you
- **Consolidated Insights** - Receive AI-generated summaries highlighting strengths, areas for improvement, and actionable next steps
- **Google OAuth Sign-In** - Secure, one-click authentication

---

## How It Works

**Workflow:**
1. **Admin creates feedback round** - Select subject, assign reviewers, set deadline
2. **Reviewers receive notification** - Submit anonymous feedback via structured form
3. **Round closes automatically** - No more submissions after deadline
4. **AI consolidates feedback** - Gemini generates insights preserving anonymity
5. **Admin reviews & shares** - Add optional notes before sharing with subject
6. **Subject receives feedback** - Access consolidated insights for professional growth

**4 Standard Questions:**
- What are this person's key strengths?
- What areas could they improve?
- What should they continue doing?
- What should they start or stop doing?

---

## Current Status

### Recently Completed
✅ **Admin Analytics** - Org-wide dashboard: counters, status donut, completion-rate trend, per-team activity, top theme extraction
✅ **Phosphor Icon System** - Consistent icons across the app
✅ **Personal Analytics** - Radar chart + strengths/improvements/insights trend on the user dashboard
✅ **PDF Export** - Branded download of consolidated feedback
✅ **Backend Test Pyramid** - Unit, in-memory integration, and real-MongoDB gateway tests
✅ **Security Hardening** - Dev-login + debug endpoints gated by `DEV_MODE` / `AdminOnly`
✅ **Dockerized Stack** - One-command setup via `docker-compose`
✅ **Teams Feature** - Team admins manage feedback within their team
✅ **Audit Logging** - Status transitions tracked end-to-end
✅ **AI Integration** - Google Gemini for consolidation
✅ **Anonymous Feedback System** - Core rounds and submission flow

### Production Ready
- Google OAuth authentication
- Role-based access control (Admin / Team Admin / Member)
- Full CRUD operations for rounds and feedback
- Real-time dashboard updates
- AI-generated consolidated feedback
- PDF export of shared consolidations
- Personal analytics for users

---

## Technology Stack

**Backend:** Go + MongoDB + Google Gemini API
**Frontend:** Vue.js 3 + TypeScript
**Authentication:** Google OAuth 2.0
**Deployment:** Docker-ready, cloud-agnostic

---

## User Roles

| Role | Capabilities |
|------|-------------|
| **Admin** | Create rounds, assign reviewers, view all feedback, generate consolidations, manage users, manage teams, share feedback |
| **Team Admin** | Manage members within their team, create team rounds, view team feedback |
| **Member** | Submit feedback, view assigned reviews, access received feedback (with PDF export and personal analytics) |

---

## Value Proposition

1. **Psychological Safety** - Anonymity encourages honest, constructive feedback
2. **AI-Powered Efficiency** - Gemini reduces admin burden by 80% vs manual consolidation
3. **Structured Process** - Consistent 4-question framework ensures quality insights
4. **Actionable Insights** - AI extracts themes and patterns humans might miss
5. **Privacy First** - No reviewer attribution, ever

---

## Metrics & Scale

- **Feedback Round Duration:** Typically 1-2 weeks
- **Typical Round Size:** 3-8 reviewers per subject
- **Consolidation Time:** < 30 seconds with Gemini
- **Anonymity:** 100% - reviewer identity never linked to responses

---

## Next Steps (Suggested)

The roadmap lives in [GitHub Issues](https://github.com/mondial7/smart-360/issues).
A few of the highest-priority items currently open:

- [Frontend automated testing (Vitest + Playwright)](https://github.com/mondial7/smart-360/issues/28)
- [360 comparison across rounds](https://github.com/mondial7/smart-360/issues/29)
- [Slack / Teams integration](https://github.com/mondial7/smart-360/issues/30)
- [Anonymous comments on shared feedback](https://github.com/mondial7/smart-360/issues/31)
- [Security hardening (CSRF, CSP, input validation)](https://github.com/mondial7/smart-360/issues/32)

---

**Project Status:** ✅ Production-ready
**Last Updated:** May 2026
**Tech Lead:** Smart 360 Engineering Team

