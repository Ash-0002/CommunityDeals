"use client";

// Create-campaign flow — /campaigns/new
// Mirrors the mobile app's create-campaign screen: pick a community, fill in
// the deal, and edit dynamic pricing tiers before publishing.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronLeft, Plus, Trash2 } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api, APIError } from "@/lib/api";
import type { CommunityResponse } from "@/lib/types";

type TierRow = { minCount: number; maxCount: number; priceRupees: number };

function toDateInputValue(d: Date): string {
  return d.toISOString().slice(0, 10);
}

export default function CreateCampaignPage() {
  const router = useRouter();

  const [communities, setCommunities] = useState<CommunityResponse[]>([]);
  const [communityId, setCommunityId] = useState("");
  const [loadingCommunities, setLoadingCommunities] = useState(true);

  const [serviceName, setServiceName] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [minParticipants, setMinParticipants] = useState("10");
  const [maxParticipants, setMaxParticipants] = useState("");
  const [serviceDate, setServiceDate] = useState(
    toDateInputValue(new Date(Date.now() + 5 * 86_400_000)),
  );
  const [endDate, setEndDate] = useState(toDateInputValue(new Date(Date.now() + 3 * 86_400_000)));

  const [tiers, setTiers] = useState<TierRow[]>([
    { minCount: 1, maxCount: 9, priceRupees: 999 },
    { minCount: 10, maxCount: 0, priceRupees: 499 },
  ]);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    api.communities
      .myCommunities({ limit: 50 })
      .then((res) => {
        setCommunities(res.communities);
        if (res.communities[0]) setCommunityId(res.communities[0].id);
      })
      .catch(() => {
        /* handled by the empty-state below */
      })
      .finally(() => setLoadingCommunities(false));
  }, []);

  function updateTier(index: number, patch: Partial<TierRow>) {
    setTiers((prev) => prev.map((t, i) => (i === index ? { ...t, ...patch } : t)));
  }

  function addTier() {
    setTiers((prev) => {
      const last = prev[prev.length - 1]!;
      const withCappedLast = prev.map((t, i) =>
        i === prev.length - 1 ? { ...t, maxCount: last.minCount + 9 } : t,
      );
      return [
        ...withCappedLast,
        {
          minCount: last.maxCount === 0 ? last.minCount + 10 : last.maxCount + 1,
          maxCount: 0,
          priceRupees: Math.round(last.priceRupees * 0.8),
        },
      ];
    });
  }

  function removeTier(index: number) {
    setTiers((prev) => (prev.length > 1 ? prev.filter((_, i) => i !== index) : prev));
  }

  function validateTiers(): string | null {
    for (let i = 1; i < tiers.length; i++) {
      if (tiers[i]!.minCount <= tiers[i - 1]!.minCount) {
        return `Tier ${i + 1}'s "from" count must be greater than the tier above it`;
      }
      if (tiers[i]!.priceRupees > tiers[i - 1]!.priceRupees) {
        return `Tier ${i + 1}'s price should be equal to or lower than the tier above it`;
      }
    }
    if (tiers[0]!.minCount !== 1) return "The first tier must start from 1 person";
    return null;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!communityId) {
      setError("Choose a community first");
      return;
    }
    const tierError = validateTiers();
    if (tierError) {
      setError(tierError);
      return;
    }
    const minP = parseInt(minParticipants, 10);
    if (!minP || minP < 1) {
      setError("Enter a valid minimum participant count");
      return;
    }

    setSubmitting(true);
    try {
      const campaign = await api.campaigns.create({
        community_id: communityId,
        service_name: serviceName.trim(),
        title: title.trim(),
        description: description.trim(),
        image_url: imageUrl.trim() || undefined,
        min_participants: minP,
        max_participants: maxParticipants ? parseInt(maxParticipants, 10) : 0,
        service_date: new Date(`${serviceDate}T10:00:00`).toISOString(),
        start_date: new Date().toISOString(),
        end_date: new Date(`${endDate}T23:59:59`).toISOString(),
        pricing_tiers: tiers.map((t) => ({
          min_count: t.minCount,
          max_count: t.maxCount,
          price: t.priceRupees * 100,
        })),
      });
      router.replace(`/campaigns/${campaign.id}`);
    } catch (err) {
      setError(err instanceof APIError ? err.message : "Failed to create deal");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AppLayout>
      <button
        onClick={() => router.back()}
        className="mb-4 inline-flex items-center gap-1 rounded-full border border-border bg-card px-3.5 py-1.5 text-[13px] font-semibold text-muted transition-colors hover:text-ink"
      >
        <ChevronLeft className="h-4 w-4" />
        Back
      </button>

      <div className="mx-auto max-w-2xl">
        <div className="mb-6">
          <p className="text-[13px] font-semibold uppercase tracking-[0.14em] text-muted">
            New group deal
          </p>
          <h1 className="mt-1.5 text-[28px] font-extrabold leading-tight tracking-tight text-ink sm:text-[36px]">
            Get your neighbours a better price
          </h1>
        </div>

        <form
          onSubmit={handleSubmit}
          className="rounded-3xl border border-border bg-card p-5 shadow-card sm:p-7"
        >
          {!loadingCommunities && communities.length === 0 && (
            <p className="mb-4 text-sm text-muted">
              Join a community first — deals are posted inside a community.
            </p>
          )}

          {communities.length > 0 && (
            <div className="mb-4">
              <label className="mb-1.5 block text-sm font-medium text-ink">Community</label>
              <select
                value={communityId}
                onChange={(e) => setCommunityId(e.target.value)}
                className="w-full rounded-lg border border-border bg-card px-4 py-2.5 text-sm text-ink focus:outline-none focus:ring-2 focus:ring-primary-500/40"
              >
                {communities.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="mb-4">
            <Input
              label="Service (e.g. AC Service)"
              value={serviceName}
              onChange={(e) => setServiceName(e.target.value)}
              required
            />
          </div>
          <div className="mb-4">
            <Input
              label="Deal title (e.g. AC Service This Sunday)"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>
          <div className="mb-4">
            <label className="mb-1.5 block text-sm font-medium text-ink">Description (optional)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-border bg-card px-4 py-2.5 text-sm text-ink placeholder-muted/70 focus:outline-none focus:ring-2 focus:ring-primary-500/40"
            />
          </div>
          <div className="mb-4">
            <Input
              label="Image URL (optional)"
              placeholder="https://…"
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
            />
          </div>

          <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Input
              label="Min. participants"
              type="number"
              min={1}
              value={minParticipants}
              onChange={(e) => setMinParticipants(e.target.value)}
              required
            />
            <Input
              label="Max (optional, 0 = no limit)"
              type="number"
              min={0}
              value={maxParticipants}
              onChange={(e) => setMaxParticipants(e.target.value)}
            />
          </div>

          <div className="mb-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Input
              label="Service date"
              type="date"
              value={serviceDate}
              onChange={(e) => setServiceDate(e.target.value)}
              required
            />
            <Input
              label="Joining closes"
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              required
            />
          </div>

          <div className="mb-2 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-ink">Dynamic pricing</h2>
            <button
              type="button"
              onClick={addTier}
              className="flex items-center gap-1 text-xs font-semibold text-primary-600 hover:text-primary-700"
            >
              <Plus className="h-3.5 w-3.5" /> Add tier
            </button>
          </div>
          <p className="mb-3 text-xs text-muted">
            Price drops automatically as more people join. The first tier must start from 1.
          </p>

          <div className="mb-6 flex flex-col gap-2">
            {tiers.map((tier, i) => (
              <div key={i} className="flex flex-wrap items-end gap-2 rounded-2xl border border-border p-3">
                <TierNumberInput
                  label="From"
                  value={tier.minCount}
                  onChange={(v) => updateTier(i, { minCount: v })}
                />
                <TierNumberInput
                  label="To (0=∞)"
                  value={tier.maxCount}
                  onChange={(v) => updateTier(i, { maxCount: v })}
                />
                <TierNumberInput
                  label="₹ price"
                  value={tier.priceRupees}
                  onChange={(v) => updateTier(i, { priceRupees: v })}
                />
                <button
                  type="button"
                  onClick={() => removeTier(i)}
                  disabled={tiers.length <= 1}
                  className="rounded-md p-2 text-red-500 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-30 dark:hover:bg-red-950/40"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>

          {error && <p className="mb-4 text-xs font-medium text-red-500">{error}</p>}

          <Button type="submit" fullWidth size="lg" loading={submitting}>
            Publish deal
          </Button>
        </form>
      </div>
    </AppLayout>
  );
}

function TierNumberInput({
  label,
  value,
  onChange,
}: {
  label: string;
  value: number;
  onChange: (v: number) => void;
}) {
  return (
    <div className="min-w-[80px] flex-1">
      <label className="mb-1 block text-[11px] font-medium text-muted">{label}</label>
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(parseInt(e.target.value, 10) || 0)}
        className="w-full rounded-md border border-border bg-card px-2.5 py-1.5 text-sm text-ink focus:outline-none focus:ring-2 focus:ring-primary-500/40"
      />
    </div>
  );
}
