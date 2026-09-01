-- Migration: 000003_create_campaigns_table (DOWN)
-- Drops in reverse dependency order:
--   participants → pricing_tiers → campaigns → triggers/functions

DROP TRIGGER  IF EXISTS set_campaign_participants_updated_at ON campaign_participants;
DROP TRIGGER  IF EXISTS set_campaigns_updated_at             ON campaigns;

DROP TABLE IF EXISTS campaign_participants  CASCADE;
DROP TABLE IF EXISTS campaign_pricing_tiers CASCADE;
DROP TABLE IF EXISTS campaigns              CASCADE;

-- Only drop the shared trigger function if no other table uses it.
-- (migrations 000001/000002 do not create it, so it is safe to drop here.)
DROP FUNCTION IF EXISTS trigger_set_updated_at();
