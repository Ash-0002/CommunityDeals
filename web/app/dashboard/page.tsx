"use client";

// User dashboard — /dashboard
// Distinct from /campaigns (the full catalog): this surfaces one spotlight
// deal (closest to unlocking, not yet joined) plus quick stats, rather than
// repeating the same list.

import { useEffect, useState } from "react";
import Link from "next/link";
import { Users, ArrowRight } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { CampaignCard } from "@/components/campaign/campaign-card";
import { CampaignHeroBanner, JoinedAvatarsRow, DiscountBadge } from "@/components/campaign/campaign-visuals";
import { ProgressBar } from "@/components/ui/progress-bar";
import { api, APIError } from "@/lib/api";
import { formatPaise, daysLeft } from "@/lib/format";
import { useUser } from "@/lib/use-user";
import type { CampaignListItem, CommunityResponse } from "@/lib/types";

export default function DashboardPage() {
  const { user } = useUser();
  const [communities, setCommunities] = useState<CommunityResponse[]>([]);
  const [campaigns, setCampaigns] = useState<CampaignListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([
      api.communities.myCommunities({ limit: 10 }),
      api.campaigns.list({ status: "PUBLISHED", limit: 20 }),
      api.campaigns.list({ status: "MINIMUM_REACHED", limit: 20 }),
    ])
      .then(([comm, camp1, camp2]) => {
        if (cancelled) return;
        setCommunities(comm.communities);
        setCampaigns([...camp1.campaigns, ...camp2.campaigns]);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof APIError ? err.message : "Failed to load dashboard");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const spotlight = pickSpotlight(campaigns);
  const rest = campaigns.filter((c) => c.id !== spotlight?.id);
  const joined = campaigns.filter((c) => c.is_joined);
  const saved = joined.reduce((sum, c) => sum + Math.max(0, c.first_tier_price - c.current_price), 0);

  return (
    <AppLayout>
      <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
        <div>
          <h1 className="text-base font-bold text-ink">
            Hi{user?.name ? `, ${user.name.split(" ")[0]}` : ""}
          </h1>
          <p className="text-xs text-muted">{user?.phone ?? ""}</p>
        </div>
        <Link
          href="/campaigns/new"
          className="rounded-lg bg-primary-600 px-3.5 py-1.5 text-xs font-semibold text-white hover:bg-primary-700"
        >
          + Start a deal
        </Link>
      </div>

      <div className="flex-1 overflow-y-auto px-6 py-6">
        {loading ? (
          <p className="text-sm text-muted">Loading…</p>
        ) : error ? (
          <p className="text-sm text-red-500">{error}</p>
        ) : (
          <div className="grid grid-cols-[1fr_280px] items-start gap-6">
            {/* Left: spotlight + more deals */}
            <div>
              <StatsRow joinedCount={joined.length} saved={saved} communityCount={communities.length} />

              {spotlight && (
                <div className="mt-6">
                  <div className="mb-3 flex items-center justify-between">
                    <h2 className="text-sm font-semibold text-ink">Today&apos;s spotlight</h2>
                    <Link href="/campaigns" className="flex items-center gap-1 text-xs font-medium text-primary-600 hover:text-primary-700">
                      See all <ArrowRight className="h-3 w-3" />
                    </Link>
                  </div>
                  <SpotlightCard item={spotlight} />
                </div>
              )}

              <div className="mt-6">
                <h2 className="mb-3 text-sm font-semibold text-ink">More active deals</h2>
                {rest.length === 0 ? (
                  <p className="rounded-xl border border-dashed border-border p-6 text-center text-sm text-muted">
                    No other active deals right now.
                  </p>
                ) : (
                  <div className="flex flex-col gap-2.5">
                    {rest.slice(0, 3).map((c) => (
                      <CampaignCard key={c.id} campaign={c} />
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Right: my communities */}
            <div>
              <h2 className="mb-3 text-sm font-semibold text-ink">My Communities</h2>
              {communities.length === 0 ? (
                <p className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted">
                  You haven&apos;t joined a community yet.
                </p>
              ) : (
                <div className="flex flex-col gap-2.5">
                  {communities.map((c) => (
                    <div key={c.id} className="rounded-xl border border-border bg-card p-3.5">
                      <p className="text-[13px] font-semibold text-ink">{c.name}</p>
                      <p className="mt-0.5 flex items-center gap-1 text-[11px] text-muted">
                        <Users className="h-3 w-3" />
                        {c.member_count} members{c.city ? ` · ${c.city}` : ""}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </AppLayout>
  );
}

/** Prefers a deal the user hasn't joined yet, closest to unlocking its next tier. */
function pickSpotlight(campaigns: CampaignListItem[]): CampaignListItem | null {
  if (campaigns.length === 0) return null;
  const candidates = campaigns.filter((c) => !c.is_joined);
  const pool = candidates.length > 0 ? candidates : campaigns;
  return [...pool].sort((a, b) => {
    const aLeft = Math.max(0, a.min_participants - a.participant_count);
    const bLeft = Math.max(0, b.min_participants - b.participant_count);
    return aLeft - bLeft;
  })[0]!;
}

function StatsRow({
  joinedCount,
  saved,
  communityCount,
}: {
  joinedCount: number;
  saved: number;
  communityCount: number;
}) {
  return (
    <div className="grid grid-cols-3 gap-2.5">
      <StatTile label="Deals joined" value={String(joinedCount)} />
      <StatTile label="You've saved" value={formatPaise(saved)} />
      <StatTile label="Communities" value={String(communityCount)} />
    </div>
  );
}

function StatTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-card px-3 py-3.5 text-center">
      <p className="text-base font-black text-primary-700 dark:text-primary-400">{value}</p>
      <p className="mt-0.5 text-[10.5px] font-semibold text-muted">{label}</p>
    </div>
  );
}

function SpotlightCard({ item }: { item: CampaignListItem }) {
  const unlocked = item.min_participants > 0 && item.participant_count >= item.min_participants;
  const progress = item.min_participants === 0 ? 0 : item.participant_count / item.min_participants;
  const percentOff =
    item.first_tier_price > item.current_price
      ? Math.round(((item.first_tier_price - item.current_price) / item.first_tier_price) * 100)
      : 0;
  const ending = daysLeft(item.end_date);

  return (
    <Link href={`/campaigns/${item.id}`} className="block">
      <div className="rounded-2xl border border-border bg-card p-4 transition-colors hover:border-primary-600/40">
        <CampaignHeroBanner imageUrl={item.image_url} label="Society announcement" />
        <p className="mt-3 text-lg font-black text-ink">{item.title}</p>
        <p className="text-sm text-muted">{item.service_name}</p>

        <div className="mt-3">
          <JoinedAvatarsRow participants={[]} totalJoined={item.participant_count} unlocked={unlocked} />
        </div>
        <div className="mt-3">
          <DiscountBadge originalPrice={item.first_tier_price} currentPrice={item.current_price} percentOff={percentOff} />
        </div>
        <div className="mt-3">
          <ProgressBar value={progress} height={10} />
        </div>
        <p className="mt-1.5 text-xs font-semibold text-muted">
          {item.participant_count} of {item.min_participants} joined{ending ? ` · ${ending}` : ""}
        </p>
      </div>
    </Link>
  );
}
