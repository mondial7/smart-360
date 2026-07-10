package repo

// Compile-time checks that every Postgres and fake type satisfies its interface.
var (
	_ UserRepository          = (*pgUsers)(nil)
	_ TeamRepository          = (*pgTeams)(nil)
	_ RoundRepository         = (*pgRounds)(nil)
	_ SubmissionRepository    = (*pgSubmissions)(nil)
	_ TemplateRepository      = (*pgTemplates)(nil)
	_ ConsolidationRepository = (*pgConsolidations)(nil)
	_ AuditRepository         = (*pgAudit)(nil)
	_ ModerationRepository    = (*pgModeration)(nil)
	_ SessionRepository       = (*pgSessions)(nil)

	_ UserRepository          = (*FakeUsers)(nil)
	_ TeamRepository          = (*FakeTeams)(nil)
	_ RoundRepository         = (*FakeRounds)(nil)
	_ SubmissionRepository    = (*FakeSubmissions)(nil)
	_ TemplateRepository      = (*FakeTemplates)(nil)
	_ ConsolidationRepository = (*FakeConsolidations)(nil)
	_ AuditRepository         = (*FakeAudit)(nil)
	_ ModerationRepository    = (*FakeModeration)(nil)
	_ SessionRepository       = (*FakeSessions)(nil)
)
