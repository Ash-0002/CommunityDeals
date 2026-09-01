-- Migration: 000003_create_campaigns_table (UP)
-- Creates the campaigns, campaign_pricing_tiers, and campaign_participants tables.
--
-- Design notes:
--   • participant_count is an INTEGER column (not a subquery) — we maintain it
--     with a SELECT FOR UPDATE in the JoinCampaign / LeaveCampaign repository
--     methods so it stays fast and race-condition-free.
--   • All prices are stored in PAISE (int64). Never use NUMERIC/FLOAT for money.
--   • campaign_pricing_tiers.tier_order is the sort key (lowest to highest
--     min_count). The service validates that prices strictly decrease.
--   • campaign_participants has a UNIQUE(campaign_id, user_id) constraint — the
--     repository also uses SELECT FOR UPDATE so this acts as a second safety net.
--   • Slugs are unique across the table and indexed for fast public-URL lookups.

-- ─────────────────────────────────────────────────────────────
-- 1. campaigns
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS campaigns (
    id                  TEXT        PRIMARY KEY,
    community_id        TEXT        NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    created_by_id       TEXT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    title               TEXT        NOT NULL,
    description         TEXT        NOT NULL DEFAULT '',
    service_type        TEXT        NOT NULL,     -- e.g. "car_wash", "ac_servicing"
    slug                TEXT        NOT NULL UNIQUE,

    status              TEXT        NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN (
                                'DRAFT',
                                'PUBLISHED',
                                'MINIMUM_REACHED',
                                'CONFIRMED',
                                'PAYMENT_PENDING',
                                'PAYMENT_COMPLETED',
                                'IN_PROGRESS',
                                'COMPLETED',
                                'CANCELLED',
                                'EXPIRED'
                            )),

    -- Participation limits
    min_participants    INTEGER     NOT NULL DEFAULT 1 CHECK (min_participants >= 1),
    max_participants    INTEGER                        CHECK (max_participants IS NULL OR max_participants >= min_participants),
    participant_count   INTEGER     NOT NULL DEFAULT 0 CHECK (participant_count >= 0),

    -- Scheduling
    service_date        TIMESTAMPTZ,
    campaign_end_date   TIMESTAMPTZ,               -- when the campaign stops accepting joins
    registration_deadline TIMESTAMPTZ,

    -- Location
    location_name       TEXT        NOT NULL DEFAULT '',
    location_address    TEXT        NOT NULL DEFAULT '',
    city                TEXT        NOT NULL DEFAULT '',
    pin_code            TEXT        NOT NULL DEFAULT '',

    -- Media & meta
    image_url           TEXT        NOT NULL DEFAULT '',
    vendor_name         TEXT        NOT NULL DEFAULT '',
    vendor_id           TEXT,                      -- FK to vendors table (added in Vendor milestone)
    terms_and_conditions TEXT       NOT NULL DEFAULT '',
    notes               TEXT        NOT NULL DEFAULT '',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fast lookups for: community campaign lists, status filtering, slug public URLs
CREATE INDEX IF NOT EXISTS idx_campaigns_community_id  ON campaigns(community_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_status        ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_slug          ON campaigns(slug);
CREATE INDEX IF NOT EXISTS idx_campaigns_created_by    ON campaigns(created_by_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_service_date  ON campaigns(service_date);
CREATE INDEX IF NOT EXISTS idx_campaigns_end_date      ON campaigns(campaign_end_date);

-- ─────────────────────────────────────────────────────────────
-- 2. campaign_pricing_tiers
-- ─────────────────────────────────────────────────────────────
-- Each row is one price break-point. Example for a car-wash campaign:
--   tier_order=1  min_count=1   price=60000  (₹600 for 1–9 people)
--   tier_order=2  min_count=10  price=50000  (₹500 for 10–24 people)
--   tier_order=3  min_count=25  price=40000  (₹400 for 25+ people)
CREATE TABLE IF NOT EXISTS campaign_pricing_tiers (
    id              TEXT        PRIMARY KEY,
    campaign_id     TEXT        NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    min_count       INTEGER     NOT NULL CHECK (min_count >= 1),
    price           BIGINT      NOT NULL CHECK (price > 0),  -- in paise
    label           TEXT        NOT NULL DEFAULT '',         -- e.g. "Early bird", "Group deal"
    tier_order      INTEGER     NOT NULL,                    -- 1-based sort order

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (campaign_id, min_count),
    UNIQUE (campaign_id, tier_order)
);

CREATE INDEX IF NOT EXISTS idx_pricing_tiers_campaign ON campaign_pricing_tiers(campaign_id, tier_order);

-- ─────────────────────────────────────────────────────────────
-- 3. campaign_participants
-- ─────────────────────────────────────────────────────────────
-- Records each user's participation in a campaign. A user may appear at most
-- once per campaign (UNIQUE constraint + repo-level SELECT FOR UPDATE).
-- When a user leaves, status → 'LEFT' and left_at is set; the row is kept for
-- analytics. On a re-join a new row is inserted (previous LEFT row stays).
CREATE TABLE IF NOT EXISTS campaign_participants (
    id              TEXT        PRIMARY KEY,
    campaign_id     TEXT        NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id         TEXT        NOT NULL REFERENCES users(id)     ON DELETE CASCADE,

    status          TEXT        NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('ACTIVE', 'LEFT', 'PAYMENT_PENDING', 'PAYMENT_COMPLETED', 'CANCELLED')),

    -- Price locked at the moment the user joined (in paise). This never
    -- changes once set — the user always pays what was shown when they joined.
    price_locked    BIGINT      NOT NULL DEFAULT 0 CHECK (price_locked >= 0),

    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at         TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Only one ACTIVE participation per user per campaign at a time.
    -- A user who left (status=LEFT) does not block a future re-join.
    UNIQUE (campaign_id, user_id, status)
    -- Note: UNIQUE(campaign_id, user_id) alone would prevent re-join after
    -- leaving, so we scope uniqueness to (campaign_id, user_id, status).
    -- The repo enforces the "no duplicate ACTIVE row" rule via SELECT FOR UPDATE.
);

CREATE INDEX IF NOT EXISTS idx_participants_campaign  ON campaign_participants(campaign_id);
CREATE INDEX IF NOT EXISTS idx_participants_user      ON campaign_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_participants_status    ON campaign_participants(campaign_id, status);

-- ─────────────────────────────────────────────────────────────
-- 4. updated_at auto-trigger (campaigns)
-- ─────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE TRIGGER set_campaigns_updated_at
    BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_campaign_participants_updated_at
    BEFORE UPDATE ON campaign_participants
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
