-- Track whether a user has completed the first-login onboarding tour.
-- NULL means "not yet onboarded" → the tour is shown on next login.
ALTER TABLE users ADD COLUMN onboarded_at timestamptz;
