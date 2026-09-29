"use client";

// Profile — /profile
// View + edit the signed-in user's name/email, see their communities (with
// location), and log out. Log out also lives in the top-bar avatar menu.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { MapPin, Users } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api, APIError } from "@/lib/api";
import { useUser, userInitials } from "@/lib/use-user";
import { getRefreshToken, clearTokens, saveUser } from "@/lib/auth";
import { shortLocation } from "@/lib/format";
import type { CommunityResponse } from "@/lib/types";

export default function ProfilePage() {
  const router = useRouter();
  const { user } = useUser();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [communities, setCommunities] = useState<CommunityResponse[]>([]);

  useEffect(() => {
    if (user) {
      setName(user.name);
      setEmail(user.email ?? "");
    }
  }, [user]);

  useEffect(() => {
    let cancelled = false;
    api.communities
      .myCommunities({ limit: 20 })
      .then((res) => {
        if (!cancelled) setCommunities(res.communities);
      })
      .catch(() => {
        /* non-critical — the section just stays empty */
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setMessage("");
    setError("");
    try {
      const updated = await api.auth.updateProfile({ name, email });
      saveUser(updated);
      setMessage("Profile updated");
    } catch (err) {
      setError(err instanceof APIError ? err.message : "Failed to update profile");
    } finally {
      setSaving(false);
    }
  }

  async function handleLogout() {
    const refreshToken = getRefreshToken();
    try {
      if (refreshToken) await api.auth.logout(refreshToken);
    } catch {
      // best-effort
    } finally {
      clearTokens();
      router.replace("/auth/login");
    }
  }

  return (
    <AppLayout>
      {/* Identity banner */}
      <div className="mb-6 overflow-hidden rounded-3xl border border-border bg-gradient-to-br from-primary-600 to-primary-800 p-6 text-white shadow-card sm:p-8">
        <div className="flex flex-col items-center gap-4 text-center sm:flex-row sm:text-left">
          <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-full bg-white/20 text-2xl font-black">
            {userInitials(user?.name)}
          </div>
          <div className="min-w-0">
            <h1 className="truncate text-[26px] font-extrabold leading-tight tracking-tight sm:text-[32px]">
              {user?.name || "CommunityDeals user"}
            </h1>
            <p className="num mt-1 text-sm font-medium text-white/80">{user?.phone ?? ""}</p>
            {user?.email && <p className="truncate text-sm text-white/70">{user.email}</p>}
          </div>
        </div>
      </div>

      <div className="grid items-start gap-6 lg:grid-cols-2">
        {/* Edit */}
        <form
          onSubmit={handleSave}
          className="flex flex-col gap-4 rounded-3xl border border-border bg-card p-5 shadow-card sm:p-6"
        >
          <h2 className="text-[11px] font-extrabold uppercase tracking-[0.12em] text-muted">
            Your details
          </h2>
          <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} />
          <Input
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          {error && <p className="text-xs font-semibold text-red-500">{error}</p>}
          {message && <p className="text-xs font-semibold text-primary-600">{message}</p>}
          <Button type="submit" loading={saving}>
            Save changes
          </Button>
        </form>

        <div className="flex flex-col gap-6">
          {/* Communities */}
          <section className="rounded-3xl border border-border bg-card p-5 shadow-card sm:p-6">
            <h2 className="mb-3 text-[11px] font-extrabold uppercase tracking-[0.12em] text-muted">
              Your communities
            </h2>
            {communities.length === 0 ? (
              <p className="text-sm text-muted">You haven&apos;t joined a community yet.</p>
            ) : (
              <ul className="flex flex-col gap-2.5">
                {communities.map((c) => {
                  const place = shortLocation(c);
                  return (
                    <li
                      key={c.id}
                      className="rounded-2xl border border-border p-3.5"
                    >
                      <p className="truncate text-[14px] font-bold text-ink">{c.name}</p>
                      <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-[12px] text-muted">
                        {place && (
                          <span className="flex items-center gap-1">
                            <MapPin className="h-3.5 w-3.5" />
                            {place}
                          </span>
                        )}
                        <span className="num flex items-center gap-1">
                          <Users className="h-3.5 w-3.5" />
                          {c.member_count} members
                        </span>
                      </div>
                    </li>
                  );
                })}
              </ul>
            )}
          </section>

          <Button variant="danger" fullWidth size="lg" onClick={handleLogout}>
            Log out
          </Button>
        </div>
      </div>
    </AppLayout>
  );
}
