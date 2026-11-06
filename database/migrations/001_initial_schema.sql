-- =============================================================================
-- Treewiki Master Production Database Schema
-- Version: 1.0.0
-- Dialect: PostgreSQL 16+
-- NOTE: every statement is idempotent; the gateway applies this file on boot.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- -----------------------------------------------------------------------------
-- 1. Users & Authentication
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    role VARCHAR(32) NOT NULL DEFAULT 'REGISTERED' CHECK (role IN (
        'ANONYMOUS', 'REGISTERED', 'AUTOCONFIRMED', 'ROLLBACKER',
        'EDITOR', 'SYSOP', 'OVERSIGHT', 'BUREAUCRAT', 'SYSTEM_ADMIN'
    )),
    edit_count INT NOT NULL DEFAULT 0,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    profile_bio TEXT,
    avatar_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    ip_address INET NOT NULL,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions (expires_at);

-- -----------------------------------------------------------------------------
-- 2. Namespaces & Pages
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS namespaces (
    id INT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    canonical_name VARCHAR(64) NOT NULL,
    is_talk BOOLEAN NOT NULL DEFAULT FALSE,
    is_searchable BOOLEAN NOT NULL DEFAULT TRUE,
    default_protection VARCHAR(32) NOT NULL DEFAULT 'all'
);

INSERT INTO namespaces (id, name, canonical_name, is_talk, is_searchable, default_protection) VALUES
    (0, 'Main', '', FALSE, TRUE, 'all'),
    (1, 'Talk', 'Talk', TRUE, TRUE, 'all'),
    (2, 'User', 'User', FALSE, TRUE, 'all'),
    (3, 'User_talk', 'User_talk', TRUE, TRUE, 'all'),
    (4, 'Treewiki', 'Treewiki', FALSE, TRUE, 'editor_only'),
    (5, 'Treewiki_talk', 'Treewiki_talk', TRUE, TRUE, 'all'),
    (6, 'File', 'File', FALSE, TRUE, 'all'),
    (7, 'File_talk', 'File_talk', TRUE, TRUE, 'all'),
    (10, 'Template', 'Template', FALSE, TRUE, 'editor_only'),
    (11, 'Template_talk', 'Template_talk', TRUE, TRUE, 'all'),
    (12, 'Help', 'Help', FALSE, TRUE, 'editor_only'),
    (13, 'Help_talk', 'Help_talk', TRUE, TRUE, 'all'),
    (14, 'Category', 'Category', FALSE, TRUE, 'all'),
    (15, 'Category_talk', 'Category_talk', TRUE, TRUE, 'all'),
    (20, 'TopicTree', 'TopicTree', FALSE, TRUE, 'editor_only'),
    (21, 'TopicTree_talk', 'TopicTree_talk', TRUE, TRUE, 'all')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    namespace_id INT NOT NULL REFERENCES namespaces(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    is_redirect BOOLEAN NOT NULL DEFAULT FALSE,
    redirect_target_title VARCHAR(255),
    current_revision_id UUID, -- Foreign key added after revisions table
    protection_level VARCHAR(32) NOT NULL DEFAULT 'all' CHECK (protection_level IN (
        'all', 'autoconfirmed_only', 'editor_only', 'sysop_only', 'system_locked'
    )),
    view_count BIGINT NOT NULL DEFAULT 0,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    tsv_content TSVECTOR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (namespace_id, title)
);

CREATE INDEX IF NOT EXISTS idx_pages_namespace_slug ON pages (namespace_id, slug);
CREATE INDEX IF NOT EXISTS idx_pages_title_trgm ON pages USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_pages_tsv_content ON pages USING gin (tsv_content);
CREATE INDEX IF NOT EXISTS idx_pages_is_deleted ON pages (is_deleted);

-- -----------------------------------------------------------------------------
-- 3. Revisions & Storage
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS revisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    parent_revision_id UUID REFERENCES revisions(id),
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    author_ip INET,
    author_ip_masked VARCHAR(64),
    summary VARCHAR(500) NOT NULL DEFAULT '',
    byte_length INT NOT NULL,
    byte_delta INT NOT NULL DEFAULT 0,
    sha256_hash CHAR(64) NOT NULL,
    is_minor BOOLEAN NOT NULL DEFAULT FALSE,
    is_rollback BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    is_oversight_suppressed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_revisions_page_id_created ON revisions (page_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_revisions_author_id ON revisions (author_id);
CREATE INDEX IF NOT EXISTS idx_revisions_created_at ON revisions (created_at DESC);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pages_current_revision') THEN
        ALTER TABLE pages ADD CONSTRAINT fk_pages_current_revision
            FOREIGN KEY (current_revision_id) REFERENCES revisions(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS revision_text (
    revision_id UUID PRIMARY KEY REFERENCES revisions(id) ON DELETE CASCADE,
    raw_wikitext TEXT NOT NULL,
    parsed_html_cache TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 4. Page Links (Backlinks / WhatLinksHere)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS page_links (
    from_page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    to_namespace_id INT NOT NULL REFERENCES namespaces(id),
    to_page_title VARCHAR(255) NOT NULL,
    PRIMARY KEY (from_page_id, to_namespace_id, to_page_title)
);

CREATE INDEX IF NOT EXISTS idx_page_links_target ON page_links (to_namespace_id, to_page_title);

-- -----------------------------------------------------------------------------
-- 5. Categories & Taxonomy
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(128) NOT NULL UNIQUE,
    slug VARCHAR(128) NOT NULL UNIQUE,
    description TEXT,
    parent_category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS page_categories (
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (page_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_page_categories_cat ON page_categories (category_id);

-- -----------------------------------------------------------------------------
-- 6. Topic Tree & Knowledge Graph DAG
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS topic_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(128) NOT NULL,
    slug VARCHAR(128) NOT NULL UNIQUE,
    description TEXT,
    depth INT NOT NULL DEFAULT 0,
    icon VARCHAR(64) DEFAULT 'book-open',
    color_hex VARCHAR(7) DEFAULT '#0F766E',
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS topic_edges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_node_id UUID NOT NULL REFERENCES topic_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES topic_nodes(id) ON DELETE CASCADE,
    relationship_type VARCHAR(32) NOT NULL DEFAULT 'subtopic' CHECK (
        relationship_type IN ('prerequisite', 'subtopic', 'application', 'related')
    ),
    weight NUMERIC(4,2) NOT NULL DEFAULT 1.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (source_node_id, target_node_id, relationship_type)
);

CREATE TABLE IF NOT EXISTS page_topic_nodes (
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES topic_nodes(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (page_id, node_id)
);

-- -----------------------------------------------------------------------------
-- 7. Threaded Talk & Discussion System
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS talk_threads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    author_ip INET,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN' CHECK (
        status IN ('OPEN', 'UNDER_DISCUSSION', 'RFC', 'CLOSED', 'RESOLVED')
    ),
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    reply_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_talk_threads_page ON talk_threads (page_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS talk_replies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    thread_id UUID NOT NULL REFERENCES talk_threads(id) ON DELETE CASCADE,
    parent_reply_id UUID REFERENCES talk_replies(id) ON DELETE SET NULL,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    author_ip INET,
    content TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_talk_replies_thread ON talk_replies (thread_id, created_at ASC);

CREATE TABLE IF NOT EXISTS talk_polls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    thread_id UUID NOT NULL REFERENCES talk_threads(id) ON DELETE CASCADE,
    question VARCHAR(255) NOT NULL,
    quorum INT NOT NULL DEFAULT 5,
    expires_at TIMESTAMPTZ NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS talk_poll_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    poll_id UUID NOT NULL REFERENCES talk_polls(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    choice VARCHAR(16) NOT NULL CHECK (choice IN ('PRO', 'CON', 'ABSTAIN')),
    ip_address INET NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (poll_id, user_id)
);

-- -----------------------------------------------------------------------------
-- 8. Media Assets & Files
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS media_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id UUID REFERENCES users(id) ON DELETE SET NULL,
    filename VARCHAR(255) NOT NULL UNIQUE,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(128) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    s3_key VARCHAR(512) NOT NULL,
    sha256_hash CHAR(64) NOT NULL,
    license_type VARCHAR(64) NOT NULL DEFAULT 'CC_BY_SA_4_0',
    width INT,
    height INT,
    is_quarantined BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_assets_hash ON media_assets (sha256_hash);

-- -----------------------------------------------------------------------------
-- 9. Moderation, Blocks & Abuse Filters
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    target_ip INET,
    target_ip_cidr CIDR,
    blocked_by_id UUID NOT NULL REFERENCES users(id),
    reason TEXT NOT NULL,
    expires_at TIMESTAMPTZ, -- NULL means indefinite
    is_sitewide BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_blocks_user ON blocks (target_user_id);
CREATE INDEX IF NOT EXISTS idx_blocks_ip ON blocks (target_ip);

CREATE TABLE IF NOT EXISTS abuse_filters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    pattern_regex TEXT NOT NULL,
    action VARCHAR(32) NOT NULL DEFAULT 'TAG' CHECK (
        action IN ('TAG', 'WARN', 'DISALLOW', 'BLOCK')
    ),
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS abuse_filter_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filter_id UUID NOT NULL REFERENCES abuse_filters(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET NOT NULL,
    page_title VARCHAR(255) NOT NULL,
    matched_snippet TEXT,
    action_taken VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- -----------------------------------------------------------------------------
-- 10. Watchlists & System Audit Logs
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS watchlists (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    notify_email BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, page_id)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_ip INET NOT NULL,
    action_type VARCHAR(64) NOT NULL,
    target_entity_type VARCHAR(64) NOT NULL,
    target_entity_id UUID,
    details_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs (created_at DESC);

-- Full-text search support (tsvector + pg_trgm)
ALTER TABLE pages ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_pages_search_vector ON pages USING gin (search_vector);
CREATE INDEX IF NOT EXISTS idx_pages_title_trgm ON pages USING gin (title gin_trgm_ops);
