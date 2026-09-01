// Shared TypeScript types for CommunityDeals web app

export type CampaignStatus =
  | "DRAFT"
  | "PUBLISHED"
  | "MINIMUM_REACHED"
  | "CONFIRMED"
  | "COMPLETED"
  | "CANCELLED";

export type PricingTier = {
  min_count: number;
  price: number;
  label?: string;
};

export type CampaignResponse = {
  id: string;
  community_id: string;
  community_name?: string;
  created_by_id: string;
  title: string;
  description?: string;
  service_type: string;
  slug: string;
  status: CampaignStatus;
  min_participants: number;
  max_participants: number | null;
  participant_count: number;
  current_price: number;
  pricing_tiers: PricingTier[];
  next_tier: PricingTier | null;
  participants_to_next_tier: number;
  is_joined: boolean;
  service_date: string | null;
  campaign_end_date: string | null;
  location_name?: string;
  location_address?: string;
  city?: string;
  pin_code?: string;
  image_url?: string;
  vendor_name?: string;
  share_url?: string;
  created_at: string;
  updated_at: string;
};

// Alias kept for backwards compat
export type CampaignListItem = CampaignResponse;
