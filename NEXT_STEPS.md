# Smart 360 Feedback - Next Steps & Future Improvements

This document outlines potential enhancements, features, and technical improvements for Smart 360 Feedback.

> **Already shipped** — see [`PRODUCT-OVERVIEW.md`](PRODUCT-OVERVIEW.md) for the full list. Highlights: AI consolidation, role-based access, anonymous submissions, PDF export of consolidated feedback, personal analytics on the user dashboard, an admin analytics dashboard with completion trends + theme extraction, and a Phosphor-based icon system.

---

## High Priority

### 1. Frontend Type-Safety Cleanup

**Current State:** `npm run build` (which runs `vue-tsc`) reports ~29 TypeScript errors across 7 view files. Vite dev server still works because it skips type-checking; production build is blocked.

**Hot spots:**
- `CreateRoundView.vue` — `number | null` vs `string` mismatch on numeric inputs.
- `RoundDetailsView.vue` — null-safety on `round` and missing `.value` accesses.
- Several views have unused imports (`User`, `auth`, `router`, `getStatusColor`, `consRes`, `formatDateTime`, etc.).
- `ConsolidationView.vue` — unused destructured names in `v-for`.

**Estimated Effort:** Small (1 day)

---

### 2. Automated Testing — Frontend

**Current State:** Backend has a three-phase test pyramid (unit, in-memory fakes, real-MongoDB gateway). Frontend has none.

**Proposal:**
- Vitest + Vue Test Utils for component tests
- Integration tests for stores and API client
- Playwright or Cypress for end-to-end happy paths
- Add a CI workflow that runs both backend and frontend test suites on PR

**Estimated Effort:** Medium-Large (4-6 days)

---

## Medium Priority

### 3. 360 Comparison Across Rounds

Show feedback evolution over time:
- Side-by-side comparison of multiple consolidated rounds
- Trend arrows (improving / declining themes)
- Highlight persistent areas
- Personal growth tracking dashboard

**Estimated Effort:** Medium-Large (4-5 days)

---

### 4. Slack / Teams Integration

- Webhook notifications for round assignments
- Slack slash commands: `/feedback list`, `/feedback submit`
- Daily reminders posted to channels
- OAuth-based user mapping

**Estimated Effort:** Medium-Large (4-5 days)

---

### 5. Anonymous Comments on Shared Feedback

- Subjects can request clarification on consolidated feedback
- Reviewers can respond anonymously (threaded)
- Time-bound (e.g., 2 weeks after sharing)
- Admin moderation tools

**Estimated Effort:** Medium (3-4 days)

---

## Technical Improvements

### 7. Enhanced Security

- Rate limiting (`golang.org/x/time/rate` or Redis) on public + auth endpoints
- CSRF protection for state-changing operations
- Content Security Policy headers via nginx
- Input sanitization with `go-playground/validator`
- Production secrets management (Docker secrets / Vault)

**Estimated Effort:** Medium (3-4 days)

---

### 8. Performance Optimization

- Compound MongoDB indexes for the heaviest queries
- Result caching for read-heavy endpoints (Redis)
- Pagination on every list endpoint
- Frontend code splitting + image optimization
- Bundle-size review (tree-shaking)

**Estimated Effort:** Medium (2-3 days)

---

### 9. Observability & Monitoring

- Structured logging (`zap` or `logrus`) with request IDs
- Prometheus metrics + Grafana dashboards
- OpenTelemetry distributed tracing
- Alerting on error rate / latency anomalies

**Estimated Effort:** Medium-Large (4-5 days)

---

### 10. Database Backups & Disaster Recovery

- Automated MongoDB backups (daily incremental, weekly full) to cloud storage
- Point-in-time recovery
- Monthly automated restore tests
- Documented restore runbook

**Estimated Effort:** Small-Medium (1-2 days)

---

### 11. CI/CD Pipeline

GitHub Actions workflow:
- On PR: lint + run backend & frontend tests
- On merge to `main`: build Docker images, push to registry, deploy to staging
- On tag: deploy to production with auto-rollback on failure

**Estimated Effort:** Medium (2-3 days)

---

## DevOps & Infrastructure

### 12. Kubernetes Deployment

- Helm chart with Deployment / Service / Ingress manifests
- Horizontal Pod Autoscaler
- Persistent Volumes for MongoDB
- Rolling updates with health/readiness probes

**Estimated Effort:** Large (5-7 days)

---

### 13. HTTPS & Domain Setup

Reverse proxy (Caddy or nginx + Let's Encrypt) with auto-renewal, HSTS, and HTTP→HTTPS redirect. Required for OAuth in production.

**Estimated Effort:** Small (1 day)

---

### 14. Multi-Tenancy Support

Isolate data by organization, subdomain routing, per-tenant configuration, billing/usage tracking. Significant scope — only attempt if pursuing a SaaS model.

**Estimated Effort:** Very Large (10-15 days)

---

## Documentation

### 15. Contributing Guidelines

Create `CONTRIBUTING.md` covering setup, code style, PR process, issue templates, code of conduct, commit conventions.

**Estimated Effort:** Small (1 day)

---

### 16. API Documentation

Auto-generated OpenAPI spec via `swaggo/swag`, served at `/api/docs` with interactive Swagger UI.

**Estimated Effort:** Medium (2-3 days)

---

## Advanced Features

### 17. Self-Nomination for Feedback

Members can request feedback on themselves (admin approval, quota system).

**Estimated Effort:** Medium (3-4 days)

---

### 18. Peer Recognition System

"Kudos" / "shoutouts" separate from formal feedback. Public recognition, optional Slack/Teams broadcast.

**Estimated Effort:** Medium (3-4 days)

---

### 19. Manager-Specific Workflows

Manager role separate from admin: view team's consolidated feedback (with permissions), private manager notes, 1:1 feedback mode, development plan tracking.

**Estimated Effort:** Large (5-6 days)

---

## Prioritization Matrix

| Priority | Feature | Impact | Effort |
|----------|---------|--------|--------|
| 🔴 High | Frontend Type-Safety | Medium | Small |
| 🔴 High | Frontend Automated Testing | High | Medium-Large |
| 🟡 Medium | 360 Comparison | Medium | Medium-Large |
| 🟡 Medium | Enhanced Security | High | Medium |
| 🟡 Medium | CI/CD Pipeline | High | Medium |
| 🟡 Medium | Slack Integration | Medium | Medium-Large |
| 🟡 Medium | Anonymous Comments | Medium | Medium |
| 🟢 Low | Performance Optimization | Medium | Medium |
| 🟢 Low | Observability & Monitoring | Medium | Medium-Large |
| 🟢 Low | DB Backups & DR | Medium | Small-Medium |
| 🟢 Low | Kubernetes | Medium | Large |
| 🟢 Low | Multi-Tenancy | High | Very Large |

---

## Implementation Roadmap

### Phase 1: Quality (Weeks 1-2)
- Frontend type-safety cleanup
- Frontend automated testing
- CI/CD pipeline
- Enhanced security

### Phase 2: Insights (Weeks 3-4)
- 360 comparison across rounds

### Phase 3: Reach (Weeks 5-6)
- Slack / Teams integration
- Anonymous comments on shared feedback

### Phase 4: Scale (Weeks 7-9)
- Performance optimization
- Observability & monitoring
- DB backups & disaster recovery
- HTTPS / domain hardening

---

## How to Contribute

1. Check existing issues to see if someone is already working on it
2. Open an issue to discuss your approach before coding
3. Fork the repository and create a feature branch
4. Follow [`CLAUDE.md`](CLAUDE.md) for project conventions
5. Write tests for your changes
6. Submit a PR with a clear description

---

**Last Updated:** May 2026
**Project Status:** Production-ready
**Current Version:** 1.1.0
