// Typed API client for the Community Platform Go backend.
//
// Usage:
//   import { api } from "@/lib/api";
//   const campaign = await api.campaigns.getBySlug("car-wash-society");

import { getAccessToken, refreshAccessToken, clearTokens } from "@/lib/auth";
import type {
  APIResponse,
  AuthTokensResponse,
  CampaignResponse,
  CommunityResponse,
  JoinCampaignResponse,
  PaginatedCampaignsResponse,
  PaginatedCommunitiesResponse,
  ParticipantResponse,
  UserResponse,
} from "@/lib/types";

// ─────────────────────────────────────────────────────────────────────────────
// Core fetch wrapper
// ─────────────────────────────────────────────────────────────────────────────

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

class APIError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

export { APIError };

type FetchOptions = RequestInit & {
  auth?: boolean;       // attach Bearer token (default true for non-public routes)
  retry?: boolean;      // allow one token refresh + retry (default true)
};

async function request<T>(
  path: string,
  options: FetchOptions = {},
): Promise<T> {
  const { auth = true, retry = true, ...fetchOptions } = options;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(fetchOptions.headers as Record<string, string> | undefined),
  };

  if (auth) {
    const token = getAccessToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...fetchOptions,
    headers,
  });

  // 401 + we have a refresh token → try once to get a new access token
  if (res.status === 401 && retry && auth) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      return request<T>(path, { ...options, retry: false });
    }
    // refresh failed — clear tokens and throw so the caller can redirect
    clearTokens();
    throw new APIError(401, "UNAUTHORIZED", "Session expired. Please log in again.");
  }

  const body = (await res.json()) as APIResponse<T>;

  if (!body.success) {
    throw new APIError(
      res.status,
      body.error?.code ?? "UNKNOWN_ERROR",
      body.message ?? "Something went wrong",
    );
  }

  return body.data as T;
}

// ─────────────────────────────────────────────────────────────────────────────
// Auth
// ─────────────────────────────────────────────────────────────────────────────

const auth = {
  sendOTP: (phone: string) =>
    request<{ message: string }>("/api/v1/auth/send-otp", {
      method: "POST",
      body: JSON.stringify({ phone }),
      auth: false,
    }),

  verifyOTP: (phone: string, otp: string) =>
    request<AuthTokensResponse>("/api/v1/auth/verify-otp", {
      method: "POST",
      body: JSON.stringify({ phone, otp }),
      auth: false,
    }),

  refreshToken: (refreshToken: string) =>
    request<AuthTokensResponse>("/api/v1/auth/refresh-token", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
      auth: false,
    }),

  logout: (refreshToken: string) =>
    request<void>("/api/v1/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),

  getProfile: () =>
    request<UserResponse>("/api/v1/users/profile"),

  updateProfile: (data: { name?: string; email?: string }) =>
    request<UserResponse>("/api/v1/users/profile", {
      method: "PATCH",
      body: JSON.stringify(data),
    }),
};

// ─────────────────────────────────────────────────────────────────────────────
// Communities
// ─────────────────────────────────────────────────────────────────────────────

const communities = {
  list: (params?: { page?: number; limit?: number; city?: string; type?: string }) => {
    const q = new URLSearchParams();
    if (params?.page)  q.set("page",  String(params.page));
    if (params?.limit) q.set("limit", String(params.limit));
    if (params?.city)  q.set("city",  params.city);
    if (params?.type)  q.set("type",  params.type);
    return request<PaginatedCommunitiesResponse>(`/api/v1/communities?${q}`);
  },

  get: (id: string) =>
    request<CommunityResponse>(`/api/v1/communities/${id}`),

  create: (data: {
    name: string;
    description?: string;
    type: string;
    city?: string;
    pin_code?: string;
    requires_approval?: boolean;
  }) =>
    request<CommunityResponse>("/api/v1/communities", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  join: (id: string, inviteCode?: string) =>
    request<{ message: string }>(`/api/v1/communities/${id}/join`, {
      method: "POST",
      body: JSON.stringify({ invite_code: inviteCode ?? "" }),
    }),

  leave: (id: string) =>
    request<void>(`/api/v1/communities/${id}/leave`, { method: "POST" }),

  myCommunities: (params?: { page?: number; limit?: number }) => {
    const q = new URLSearchParams();
    if (params?.page)  q.set("page",  String(params.page));
    if (params?.limit) q.set("limit", String(params.limit));
    return request<PaginatedCommunitiesResponse>(`/api/v1/users/communities?${q}`);
  },
};

// ─────────────────────────────────────────────────────────────────────────────
// Campaigns
// ─────────────────────────────────────────────────────────────────────────────

const campaigns = {
  /** Public — no auth needed. Used for the shareable /c/[slug] page. */
  getBySlug: (slug: string) =>
    request<CampaignResponse>(`/c/${slug}`, { auth: false }),

  get: (id: string) =>
    request<CampaignResponse>(`/api/v1/campaigns/${id}`),

  list: (params?: {
    community_id?: string;
    status?: string;
    page?: number;
    limit?: number;
  }) => {
    const q = new URLSearchParams();
    if (params?.community_id) q.set("community_id", params.community_id);
    if (params?.status)       q.set("status",       params.status);
    if (params?.page)         q.set("page",         String(params.page));
    if (params?.limit)        q.set("limit",        String(params.limit));
    return request<PaginatedCampaignsResponse>(`/api/v1/campaigns?${q}`);
  },

  join: (id: string) =>
    request<JoinCampaignResponse>(`/api/v1/campaigns/${id}/join`, {
      method: "POST",
    }),

  leave: (id: string) =>
    request<void>(`/api/v1/campaigns/${id}/leave`, { method: "POST" }),

  participants: (id: string, limit = 50) =>
    request<{ participants: ParticipantResponse[]; total: number }>(
      `/api/v1/campaigns/${id}/participants?limit=${limit}`,
    ),

  /** Mirrors the Go backend's CreateCampaignRequest exactly (dto/campaign.go). */
  create: (data: {
    community_id: string;
    service_name: string;
    title: string;
    description?: string;
    image_url?: string;
    min_participants: number;
    max_participants?: number; // 0 = no cap
    service_date: string; // RFC3339
    start_date: string; // RFC3339
    end_date: string; // RFC3339
    pricing_tiers: Array<{ min_count: number; max_count: number; price: number }>;
  }) =>
    request<CampaignResponse>("/api/v1/campaigns", {
      method: "POST",
      body: JSON.stringify(data),
    }),
};

// ─────────────────────────────────────────────────────────────────────────────
// Unified export
// ─────────────────────────────────────────────────────────────────────────────

export const api = { auth, communities, campaigns };
