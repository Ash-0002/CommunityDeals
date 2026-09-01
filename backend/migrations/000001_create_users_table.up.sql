-- Migration: 000001 Create users and refresh_tokens tables
-- Description: Core identity tables for phone-OTP authentication

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- provides gen_random_uuid()

-- ── users ──────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    phone         VARCHAR(20) NOT NULL,
    name          VARCHAR(100) NOT NULL DEFAULT '',
    email         VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url    TEXT        NOT NULL DEFAULT '',
    role          VARCHAR(30) NOT NULL DEFAULT 'MEMBER'
                  CHECK (role IN ('MEMBER', 'COMMUNITY_ADMIN', 'VENDOR', 'PLATFORM_ADMIN')),
    status        VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                  CHECK (status IN ('ACTIVE', 'INACTIVE', 'BANNED')),
    is_verified   BOOLEAN     NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Phone must be unique (one account per number)
CREATE UNIQUE INDEX idx_users_phone ON users (phone);

-- Lookups by role for admin queries
CREATE INDEX idx_users_role ON users (role);

-- ── refresh_tokens ─────────────────────────────────────────────────────────────
-- We store only the SHA-256 hash of each refresh token, never the raw value.
CREATE TABLE refresh_tokens (
    id         TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    user_id    TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(64) NOT NULL,   -- SHA-256 hex (64 chars)
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ             -- NULL means still valid
);

-- Fast token lookup on each API call
CREATE UNIQUE INDEX idx_refresh_tokens_token ON refresh_tokens (token);

-- Efficiently find all active tokens for a user (e.g. logout-all)
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- Comment the intent of key columns
COMMENT ON TABLE  users IS 'All platform users regardless of role';
COMMENT ON COLUMN users.phone IS 'E.164 format, e.g. +919876543210';
COMMENT ON COLUMN users.role IS 'Platform-level role; community-specific roles live in community_members';

COMMENT ON TABLE  refresh_tokens IS 'Hashed refresh tokens for JWT rotation';
COMMENT ON COLUMN refresh_tokens.token IS 'SHA-256 hex of the raw refresh token issued to the client';
