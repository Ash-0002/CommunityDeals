// Token storage and refresh logic.
//
// Tokens are stored in httpOnly cookies (set server-side) for security,
// with a JS-readable copy of the access token in localStorage for the API
// client. On the server (SSR/middleware), tokens come from the request cookies.

import Cookies from "js-cookie";

const ACCESS_TOKEN_KEY  = "cp_access_token";
const REFRESH_TOKEN_KEY = "cp_refresh_token";
const USER_KEY          = "cp_user";

// ── Client-side token helpers ─────────────────────────────────────────────────

export function saveTokens(accessToken: string, refreshToken: string): void {
  // Access token in a short-lived cookie (15 min to match backend)
  Cookies.set(ACCESS_TOKEN_KEY, accessToken, {
    expires: 1 / 96, // 15 minutes as fraction of a day
    sameSite: "strict",
    secure: process.env.NODE_ENV === "production",
  });
  // Refresh token in a longer-lived cookie (7 days)
  Cookies.set(REFRESH_TOKEN_KEY, refreshToken, {
    expires: 7,
    sameSite: "strict",
    secure: process.env.NODE_ENV === "production",
  });
}

export function getAccessToken(): string | undefined {
  return Cookies.get(ACCESS_TOKEN_KEY);
}

export function getRefreshToken(): string | undefined {
  return Cookies.get(REFRESH_TOKEN_KEY);
}

export function clearTokens(): void {
  Cookies.remove(ACCESS_TOKEN_KEY);
  Cookies.remove(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

export function isLoggedIn(): boolean {
  return !!Cookies.get(ACCESS_TOKEN_KEY) || !!Cookies.get(REFRESH_TOKEN_KEY);
}

// ── User cache ────────────────────────────────────────────────────────────────

export function saveUser(user: unknown): void {
  try {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  } catch (_) {
    // ignore storage errors (private browsing, etc.)
  }
}

export function getCachedUser<T>(): T | null {
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? (JSON.parse(raw) as T) : null;
  } catch (_) {
    return null;
  }
}

// ── Token refresh ─────────────────────────────────────────────────────────────

let _refreshPromise: Promise<boolean> | null = null;

/**
 * Attempts to refresh the access token using the stored refresh token.
 * Deduplicates concurrent refresh calls — only one network request fires
 * even if multiple API calls fail with 401 simultaneously.
 *
 * Returns true if refresh succeeded, false otherwise.
 */
export async function refreshAccessToken(): Promise<boolean> {
  if (_refreshPromise) return _refreshPromise;

  _refreshPromise = (async () => {
    const refreshToken = getRefreshToken();
    if (!refreshToken) return false;

    try {
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/auth/refresh-token`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: refreshToken }),
        },
      );
      if (!res.ok) return false;

      const body = await res.json();
      if (!body.success || !body.data?.access_token) return false;

      saveTokens(body.data.access_token, body.data.refresh_token ?? refreshToken);
      if (body.data.user) saveUser(body.data.user);
      return true;
    } catch (_) {
      return false;
    } finally {
      _refreshPromise = null;
    }
  })();

  return _refreshPromise;
}
