-- =============================================================================
-- Treewiki Production Database Schema — Migration 002
-- Aligns the 001 schema with actual application runtime behavior.
-- Every statement is idempotent (safe to re-run on every gateway boot).
-- Dialect: PostgreSQL 16+
-- =============================================================================

-- Pages: soft-delete timestamp (memory Page.DeletedAt parity).
ALTER TABLE pages ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Pages: accept the protection levels the wiki service actually writes
-- ("all", "registered", "autoconfirmed", "editor", "sysop", ...).
ALTER TABLE pages DROP CONSTRAINT IF EXISTS pages_protection_level_check;
ALTER TABLE pages ADD CONSTRAINT pages_protection_level_check CHECK (protection_level IN (
    'all', 'registered', 'autoconfirmed', 'autoconfirmed_only',
    'editor', 'editor_only', 'sysop', 'sysop_only', 'system_locked'
));

-- Display-name authors (usernames, masked IPs) alongside UUID references.
ALTER TABLE revisions ADD COLUMN IF NOT EXISTS author_name TEXT;
ALTER TABLE talk_threads ADD COLUMN IF NOT EXISTS author_name TEXT;
ALTER TABLE talk_replies ADD COLUMN IF NOT EXISTS author_name TEXT;

-- Talk threads are keyed by page key ("ns:Title") so discussion stubs can
-- exist without a resolved page row; page_id stays as optional linkage.
ALTER TABLE talk_threads ALTER COLUMN page_id DROP NOT NULL;
ALTER TABLE talk_threads ADD COLUMN IF NOT EXISTS page_key TEXT;
CREATE INDEX IF NOT EXISTS idx_talk_threads_page_key ON talk_threads (page_key);
ALTER TABLE talk_threads ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE;

-- Talk replies: @mention lists and optional author IP.
ALTER TABLE talk_replies ADD COLUMN IF NOT EXISTS mentions TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE talk_replies ALTER COLUMN author_ip DROP NOT NULL;

-- Poll votes: accept the frontend canonical choices (SUPPORT/OPPOSE/NEUTRAL)
-- plus legacy schema values, and allow display-name voters.
ALTER TABLE talk_poll_votes DROP CONSTRAINT IF EXISTS talk_poll_votes_choice_check;
ALTER TABLE talk_poll_votes ADD CONSTRAINT talk_poll_votes_choice_check CHECK (
    choice IN ('PRO', 'CON', 'ABSTAIN', 'SUPPORT', 'OPPOSE', 'NEUTRAL')
);
ALTER TABLE talk_poll_votes ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE talk_poll_votes ALTER COLUMN ip_address DROP NOT NULL;
ALTER TABLE talk_poll_votes ADD COLUMN IF NOT EXISTS voter_name TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_poll_voter_name_unique
    ON talk_poll_votes (poll_id, voter_name) WHERE voter_name IS NOT NULL;

-- Abuse filter logs: display-name actors and optional IP.
ALTER TABLE abuse_filter_logs ALTER COLUMN ip_address DROP NOT NULL;
ALTER TABLE abuse_filter_logs ADD COLUMN IF NOT EXISTS actor_name TEXT;
ALTER TABLE abuse_filters ALTER COLUMN created_by_id DROP NOT NULL;

-- Blocks: display names and raw (possibly masked) IP targets.
ALTER TABLE blocks ALTER COLUMN blocked_by_id DROP NOT NULL;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS blocked_by_name TEXT;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS target_user_name TEXT;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS target_ip_raw TEXT;

-- Oversight audit ledger (legal suppression record).
CREATE TABLE IF NOT EXISTS oversight_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_name TEXT,
    action TEXT NOT NULL DEFAULT '',
    target TEXT NOT NULL DEFAULT '',
    details TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_oversight_logs_created ON oversight_logs (created_at DESC);

-- DMCA / takedown request ledger.
CREATE TABLE IF NOT EXISTS takedown_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_title VARCHAR(255) NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    requester_name TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_takedown_requests_status ON takedown_requests (status);

-- Formal dispute arbitration cases.
CREATE TABLE IF NOT EXISTS arbitration_cases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    thread_id UUID REFERENCES talk_threads(id) ON DELETE CASCADE,
    summary TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_arbitration_cases_thread ON arbitration_cases (thread_id);

-- Curated learning paths over topic nodes.
CREATE TABLE IF NOT EXISTS learning_paths (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(128) NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS learning_path_nodes (
    path_id UUID NOT NULL REFERENCES learning_paths(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES topic_nodes(id) ON DELETE CASCADE,
    position INT NOT NULL,
    PRIMARY KEY (path_id, position)
);
CREATE INDEX IF NOT EXISTS idx_learning_path_nodes_node ON learning_path_nodes (node_id);

-- Topic node application attributes (memory TopicNode parity).
ALTER TABLE topic_nodes ADD COLUMN IF NOT EXISTS difficulty INT NOT NULL DEFAULT 1;
ALTER TABLE topic_nodes ADD COLUMN IF NOT EXISTS completion DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE topic_nodes ADD COLUMN IF NOT EXISTS prereq_slugs TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE topic_nodes ADD COLUMN IF NOT EXISTS advanced_slugs TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE topic_nodes ADD COLUMN IF NOT EXISTS related_slugs TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE topic_nodes ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Media uploader display name (alongside optional UUID reference).
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS uploader_name TEXT;
