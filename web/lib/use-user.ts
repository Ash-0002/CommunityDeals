"use client";

// Small client-side hook that exposes the current user, backed by the
// localStorage cache written at login and refreshed once from the API.

import { useEffect, useState } from "react";
import { api, APIError } from "@/lib/api";
import { getCachedUser, saveUser } from "@/lib/auth";
import type { UserResponse } from "@/lib/types";

export function useUser() {
  // Start at null on both server and client's first render to avoid a
  // hydration mismatch — the localStorage cache is only readable client-side,
  // so it's applied in the effect below, after mount.
  const [user, setUser] = useState<UserResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const cached = getCachedUser<UserResponse>();
    if (cached) setUser(cached);

    let cancelled = false;
    api.auth
      .getProfile()
      .then((fresh) => {
        if (cancelled) return;
        setUser(fresh);
        saveUser(fresh);
      })
      .catch((err) => {
        if (!cancelled && err instanceof APIError) {
          // Session expired etc — leave any cached user in place, the
          // middleware will redirect to /auth/login on next navigation.
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return { user, loading };
}

export function userInitials(name: string | undefined): string {
  const trimmed = (name ?? "").trim();
  if (!trimmed) return "?";
  const parts = trimmed.split(/\s+/);
  if (parts.length === 1) return parts[0]![0]!.toUpperCase();
  return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
}
