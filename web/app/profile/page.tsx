"use client";

// Profile — /profile
// View + edit the signed-in user's name/email, and log out.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AppLayout } from "@/components/layout/app-layout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api, APIError } from "@/lib/api";
import { useUser, userInitials } from "@/lib/use-user";
import { getRefreshToken, clearTokens, saveUser } from "@/lib/auth";

export default function ProfilePage() {
  const router = useRouter();
  const { user } = useUser();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (user) {
      setName(user.name);
      setEmail(user.email ?? "");
    }
  }, [user]);

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
      <div className="border-b border-border bg-card px-6 py-4">
        <h1 className="text-base font-bold text-ink">Profile</h1>
      </div>

      <div className="flex-1 overflow-y-auto px-6 py-6">
        <div className="mx-auto max-w-md">
          <div className="mb-5 flex flex-col items-center rounded-xl border border-border bg-card p-6 text-center">
            <div className="mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-primary-600 text-xl font-bold text-white">
              {userInitials(user?.name)}
            </div>
            <p className="text-lg font-bold text-ink">{user?.name || "CommunityDeals user"}</p>
            <p className="text-sm text-muted">{user?.phone ?? ""}</p>
          </div>

          <form onSubmit={handleSave} className="mb-5 flex flex-col gap-4 rounded-xl border border-border bg-card p-5">
            <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} />
            <Input label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
            {error && <p className="text-xs font-medium text-red-500">{error}</p>}
            {message && <p className="text-xs font-medium text-primary-600">{message}</p>}
            <Button type="submit" loading={saving}>
              Save changes
            </Button>
          </form>

          <Button variant="danger" fullWidth onClick={handleLogout}>
            Log out
          </Button>
        </div>
      </div>
    </AppLayout>
  );
}
