-- Initial Smart 360 schema.
-- UUID primary keys; embedded Mongo arrays normalized into join tables
-- (team_members, round_reviewers); config/snapshot blobs kept as jsonb.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- users and teams reference each other, so users.team_id's FK is added after
-- teams exists (below).
CREATE TABLE users (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email      text NOT NULL UNIQUE,
    name       text NOT NULL DEFAULT '',
    photo_url  text NOT NULL DEFAULT '',
    role       text NOT NULL DEFAULT 'member'
                   CHECK (role IN ('admin', 'team_admin', 'member')),
    team_id    uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_login timestamptz
);

CREATE TABLE teams (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text NOT NULL,
    team_admin_id uuid NOT NULL REFERENCES users(id),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE users
    ADD CONSTRAINT users_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;

-- Membership as a join table (replaces Team.MemberIDs[]).
CREATE TABLE team_members (
    team_id    uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, user_id)
);
CREATE INDEX idx_team_members_user ON team_members(user_id);

CREATE TABLE templates (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             text NOT NULL UNIQUE,
    name             text NOT NULL,
    description      text NOT NULL DEFAULT '',
    coaching_persona text NOT NULL DEFAULT '',
    questions        jsonb NOT NULL DEFAULT '[]',
    competencies     jsonb NOT NULL DEFAULT '[]',
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feedback_rounds (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id    uuid NOT NULL REFERENCES users(id),
    created_by_id uuid NOT NULL REFERENCES users(id),
    template_id   uuid REFERENCES templates(id),
    deadline      timestamptz,
    status        text NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft', 'active', 'closed', 'shared')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_rounds_subject ON feedback_rounds(subject_id);
CREATE INDEX idx_rounds_status ON feedback_rounds(status);

-- Reviewers as a join table (replaces embedded Reviewers[]).
CREATE TABLE round_reviewers (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id    uuid NOT NULL REFERENCES feedback_rounds(id) ON DELETE CASCADE,
    reviewer_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (round_id, reviewer_id)
);
CREATE INDEX idx_round_reviewers_reviewer ON round_reviewers(reviewer_id);

CREATE TABLE submissions (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id              uuid NOT NULL REFERENCES feedback_rounds(id) ON DELETE CASCADE,
    reviewer_id           uuid NOT NULL REFERENCES users(id),
    responses             jsonb NOT NULL DEFAULT '{}',
    is_self               boolean NOT NULL DEFAULT false,
    relationship          text,
    interaction_frequency text,
    ratings               jsonb NOT NULL DEFAULT '[]',
    private_notes         text NOT NULL DEFAULT '',
    submitted_at          timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    -- one submission per reviewer per round (self counts as the subject reviewing)
    UNIQUE (round_id, reviewer_id)
);

CREATE TABLE consolidations (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id              uuid NOT NULL UNIQUE REFERENCES feedback_rounds(id) ON DELETE CASCADE,
    generated_by_id       uuid NOT NULL REFERENCES users(id),
    executive_summary     text NOT NULL DEFAULT '',
    -- formerly JSON strings in Mongo; now real jsonb
    strengths             jsonb NOT NULL DEFAULT '[]',
    areas_for_improvement jsonb NOT NULL DEFAULT '[]',
    actionable_insights   jsonb NOT NULL DEFAULT '[]',
    question_summaries    jsonb NOT NULL DEFAULT '{}',
    question_labels       jsonb NOT NULL DEFAULT '{}',
    self_vs_others_delta  jsonb,
    voice_breakdown       jsonb,
    competency_ratings    jsonb NOT NULL DEFAULT '[]',
    manager_only_channel  jsonb,
    admin_notes           text NOT NULL DEFAULT '',
    shared_at             timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- audit_logs cache display fields and are intentionally NOT foreign-keyed, so
-- the trail survives deletion of the entities it references.
CREATE TABLE audit_logs (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    action        text NOT NULL,
    actor_id      uuid NOT NULL,
    actor_name    text NOT NULL DEFAULT '',
    actor_email   text NOT NULL DEFAULT '',
    round_id      uuid,
    round_subject text NOT NULL DEFAULT '',
    team_id       uuid,
    team_name     text NOT NULL DEFAULT '',
    description   text NOT NULL DEFAULT '',
    old_value     text NOT NULL DEFAULT '',
    new_value     text NOT NULL DEFAULT '',
    metadata      text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_round ON audit_logs(round_id);

CREATE TABLE moderation_logs (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id        uuid NOT NULL,
    submission_id   uuid NOT NULL,
    model           text NOT NULL DEFAULT '',
    flagged         boolean NOT NULL DEFAULT false,
    reasons         jsonb NOT NULL DEFAULT '[]',
    fields_scrubbed jsonb NOT NULL DEFAULT '[]',
    moderated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_moderation_round_time ON moderation_logs(round_id, moderated_at DESC);

CREATE TABLE sessions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
