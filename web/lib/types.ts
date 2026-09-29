// Shared TypeScript types for the CommunityDeals web app.
// Mirrors the Go backend's response DTOs (backend/internal/dto/*.go) exactly —
// keep these two in sync when the backend shape changes.

// ── Envelope ─────────────────────────────────────────────────────────────────

export type APIError = {
  code: string;
  message: string;
  details?: unknown;
};

export type APIResponse<T> = {
  success: boolean;
  message?: string;
  data?: T;
  error?: APIError;
};

// ── Auth / users ─────────────────────────────────────────────────────────────

export type UserResponse = {
  id: string;
  phone: string;
  name: string;
  email?: string;
  avatar_url?: string;
  role: string;
  is_verified: boolean;
};

export type AuthTokensResponse = {
  access_token: string;
  refresh_token: string;
  token_type: string;
  user: UserResponse;
};

// ── Communities ──────────────────────────────────────────────────────────────

export type CommunityResponse = {
  id: string;
  name: string;
  description: string;
  type: string;
  status: string;
  city?: string;
  state?: string;
  pin_code?: string;
  address?: string;
  requires_approval: boolean;
  invite_code?: string;
  logo_url?: string;
  member_count: number;
  created_by_id: string;
  created_at: string;
};

export type PaginatedCommunitiesResponse = {
  communities: CommunityResponse[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
};

// ── Campaigns ────────────────────────────────────────────────────────────────

export type CampaignStatus =
  | "DRAFT"
  | "PUBLISHED"
  | "MINIMUM_REACHED"
  | "CONFIRMED"
  | "COMPLETED"
  | "CANCELLED";

export type PricingTierResponse = {
  min_count: number;
  max_count: number;
  price: number; // paise
  tier_order: number;
};

/** Full campaign detail — GET /campaigns/:id and GET /c/:slug */
export type CampaignResponse = {
  id: string;
  community_id: string;
  vendor_id: string;
  service_name: string;
  title: string;
  description: string;
  image_url?: string;
  min_participants: number;
  max_participants: number;
  service_date: string;
  start_date: string;
  end_date: string;
  status: CampaignStatus;
  participant_count: number;
  slug: string;
  share_url: string;
  pricing_tiers: PricingTierResponse[];
  current_price: number; // paise
  next_tier?: PricingTierResponse | null;
  participants_to_next_tier: number;
  is_joined: boolean;
  created_at: string;
};

/** Lightweight row used by GET /campaigns list endpoints */
export type CampaignListItem = {
  id: string;
  title: string;
  service_name: string;
  image_url?: string;
  status: CampaignStatus;
  participant_count: number;
  min_participants: number;
  current_price: number; // paise
  end_date: string;
  service_date: string;
  share_url: string;
  is_joined: boolean;
  first_tier_price: number; // paise — un-discounted tier 1 price
};

export type ParticipantResponse = {
  user_id: string;
  name: string;
  avatar_url?: string;
  joined_at: string;
};

export type JoinCampaignResponse = {
  participant_count: number;
  price_locked: number;
  current_tier: PricingTierResponse;
  next_tier?: PricingTierResponse | null;
  message: string;
};

export type PaginatedCampaignsResponse = {
  campaigns: CampaignListItem[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
};
