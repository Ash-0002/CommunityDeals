-- Migration: 000002 Create communities and community_members tables
-- Description: Core community grouping tables

-- ── communities ────────────────────────────────────────────────────────────────
CREATE TABLE communities (
    id                TEXT         PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    name              VARCHAR(100) NOT NULL,
    description       VARCHAR(500) NOT NULL DEFAULT '',
    type              VARCHAR(30)  NOT NULL
                      CHECK (type IN (
                          'SOCIETY', 'APARTMENT', 'RESIDENTIAL_COMPLEX',
                          'OFFICE', 'COLLEGE', 'CORPORATE_GROUP',
                          'PRIVATE_GROUP', 'PARTNER_COMMUNITY'
                      )),
    status            VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
                      CHECK (status IN ('ACTIVE', 'INACTIVE', 'ARCHIVED')),
    -- Location (optional)
    city              VARCHAR(100) NOT NULL DEFAULT '',
    state             VARCHAR(100) NOT NULL DEFAULT '',
    pin_code          VARCHAR(10)  NOT NULL DEFAULT '',
    address           VARCHAR(300) NOT NULL DEFAULT '',
    -- Settings
    requires_approval BOOLEAN      NOT NULL DEFAULT FALSE,
    invite_code       VARCHAR(10)  NOT NULL DEFAULT '',
    logo_url          TEXT         NOT NULL DEFAULT '',
    -- Ownership
    created_by_id     TEXT         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Invite code must be unique (used for quick joining)
CREATE UNIQUE INDEX idx_communities_invite_code
    ON communities (invite_code)
    WHERE invite_code != '';

-- Filter by type and location
CREATE INDEX idx_communities_type        ON communities (type);
CREATE INDEX idx_communities_city        ON communities (city);
CREATE INDEX idx_communities_pin_code    ON communities (pin_code);
CREATE INDEX idx_communities_status      ON communities (status);
CREATE INDEX idx_communities_created_by  ON communities (created_by_id);

-- ── community_members ─────────────────────────────────────────────────────────
CREATE TABLE community_members (
    id           TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    community_id TEXT        NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id      TEXT        NOT NULL REFERENCES users(id)       ON DELETE CASCADE,
    role         VARCHAR(20) NOT NULL DEFAULT 'MEMBER'
                 CHECK (role IN ('MEMBER', 'ADMIN')),
    status       VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                 CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'REMOVED')),
    joined_at    TIMESTAMPTZ,          -- set when status becomes APPROVED
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One membership record per user per community
CREATE UNIQUE INDEX idx_community_members_unique
    ON community_members (community_id, user_id);

-- Fast lookup: all communities a user belongs to
CREATE INDEX idx_community_members_user_id     ON community_members (user_id);
CREATE INDEX idx_community_members_community_id ON community_members (community_id);
-- Filter by status (e.g. list pending approvals quickly)
CREATE INDEX idx_community_members_status      ON community_members (community_id, status);

COMMENT ON TABLE  communities IS 'All platform communities — societies, apartments, offices, etc.';
COMMENT ON COLUMN communities.invite_code IS 'Short alphanumeric code for easy joining (e.g. GV4K2X)';
COMMENT ON COLUMN communities.requires_approval IS 'If true, joins create PENDING memberships until an admin approves';

COMMENT ON TABLE  community_members IS 'User membership in communities with role and approval state';
COMMENT ON COLUMN community_members.joined_at IS 'Set when status transitions to APPROVED';
