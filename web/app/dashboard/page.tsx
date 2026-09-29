"use client";

// Home / dashboard — /dashboard
//
// A bento grid rather than a list: one large spotlight tile (2×2 at lg), three
// small stat tiles, a community tile with the society's location, then a row of
// the remaining active deals. Collapses to a single column below md.

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, Handshake, MapPin, PiggyBank, Users } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { CampaignCard } from "@/components/campaign/campaign-card";
import {
  CampaignHeroBanner,
  JoinedAvatarsRow,
  DiscountBadge,
} from "@/components/campaign/campaign-visuals";
import { ProgressBar } from "@/components/ui/progress-bar";
import { api, APIError } from "@/lib/api";
import { formatPaise, daysLeft, shortLocation } from "@/lib/format";
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
  const saved = joined.reduce(
    (sum, c) => sum + Math.max(0, c.first_tier_price - c.current_price),
    0,
  );
  const home = communities[0] ?? null;

  return (
    <AppLayout>
      {/* Greeting */}
      <div className="mb-6 sm:mb-8">
        <p className="text-[13px] font-semibold uppercase tracking-[0.14em] text-muted">
          Your neighbourhood
        </p>
        <h1 className="mt-1.5 text-[30px] font-extrabold leading-[1.08] tracking-tight text-ink sm:text-[40px]">
          Hi{user?.name ? `, ${user.name.split(" ")[0]}` : " there"} 👋
          <br className="hidden sm:block" />
          <span className="text-muted"> here&apos;s what&apos;s happening.</span>
        </h1>
      </div>

      {loading ? (
        <p className="text-sm text-muted">Loading…</p>
      ) : error ? (
        <p className="text-sm text-red-500">{error}</p>
      ) : (
        <>
          {/* ── Bento ─────────────────────────────────────────────────────── */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
            {spotlight ? (
              <SpotlightTile item={spotlight} />
            ) : (
              <div className="rounded-3xl border border-dashed border-border p-10 text-center text-sm text-muted md:col-span-2 md:row-span-2">
                No active deals in your society yet.{" "}
                <Link href="/campaigns/new" className="font-semibold text-primary-600">
                  Start one →
                </Link>
              </div>
            )}

            <StatTile
              icon={Handshake}
              label="Deals joined"
              value={String(joined.length)}
              tone="green"
            />
            <StatTile
              icon={PiggyBank}
              label="You've saved"
              value={formatPaise(saved)}
              tone="amber"
            />
            <StatTile
              icon={Users}
              label="Communities"
              value={String(communities.length)}
              tone="sky"
            />

            <CommunityTile community={home} extra={Math.max(0, communities.length - 1)} />
          </div>

          {/* ── More active deals ─────────────────────────────────────────── */}
          <div className="mt-10">
            <div className="mb-4 flex items-end justify-between gap-3">
              <h2 className="text-xl font-extrabold tracking-tight text-ink sm:text-2xl">
                More active deals
              </h2>
              <Link
                href="/campaigns"
                className="flex shrink-0 items-center gap-1 text-[13px] font-bold text-primary-600 hover:text-primary-700"
              >
                See all <ArrowRight className="h-3.5 w-3.5" />
              </Link>
            </div>

            {rest.length === 0 ? (
              <p className="rounded-3xl border border-dashed border-border p-8 text-center text-sm text-muted">
                No other active deals right now.
              </p>
            ) : (
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {rest.slice(0, 6).map((c) => (
                  <CampaignCard key={c.id} campaign={c} />
                ))}
              </div>
            )}
          </div>
        </>
      )}
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

// ── Tiles ───────────────────────────────────────────────────────────────────

const TONES = {
  green: "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300",
  amber: "bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300",
  sky: "bg-sky-100 text-sky-700 dark:bg-sky-400/15 dark:text-sky-300",
} as const;

function StatTile({
  icon: Icon,
  label,
  value,
  tone,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  tone: keyof typeof TONES;
}) {
  return (
    <div className="flex flex-col justify-between rounded-3xl border border-border bg-card p-5 shadow-card">
      <span className={`flex h-9 w-9 items-center justify-center rounded-xl ${TONES[tone]}`}>
        <Icon className="h-[18px] w-[18px]" />
      </span>
      <div className="mt-6">
        <p className="num text-[30px] font-black leading-none tracking-tight text-ink">{value}</p>
        <p className="mt-1.5 text-[12px] font-semibold text-muted">{label}</p>
      </div>
    </div>
  );
}

function CommunityTile({
  community,
  extra,
}: {
  community: CommunityResponse | null;
  extra: number;
}) {
  if (!community) {
    return (
      <div className="flex flex-col justify-center rounded-3xl border border-dashed border-border p-5 text-center">
        <p className="text-[13px] font-semibold text-ink">No community yet</p>
        <p className="mt-1 text-xs text-muted">Join your society to see its deals.</p>
      </div>
    );
  }

  const place = shortLocation(community);

  return (
    <div className="flex flex-col justify-between rounded-3xl border border-border bg-gradient-to-br from-primary-600 to-primary-800 p-5 text-white shadow-card">
      <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-white/20">
        <MapPin className="h-[18px] w-[18px]" />
      </span>
      <div className="mt-6 min-w-0">
        <p className="truncate text-[17px] font-extrabold leading-tight">{community.name}</p>
        {place && <p className="mt-1 truncate text-[12.5px] font-medium text-white/80">{place}</p>}
        <p className="num mt-2.5 text-[12px] font-bold text-white/90">
          {community.member_count} members
          {extra > 0 ? ` · +${extra} more community${extra === 1 ? "" : "s"}` : ""}
        </p>
      </div>
    </div>
  );
}

function SpotlightTile({ item }: { item: CampaignListItem }) {
  const unlocked = item.min_participants > 0 && item.participant_count >= item.min_participants;
  const progress =
    item.min_participants === 0 ? 0 : item.participant_count / item.min_participants;
  const percentOff =
    item.first_tier_price > item.current_price
      ? Math.round(((item.first_tier_price - item.current_price) / item.first_tier_price) * 100)
      : 0;
  const ending = daysLeft(item.end_date);

  return (
    <div className="flex flex-col overflow-hidden rounded-3xl border border-border bg-card shadow-card transition-colors hover:border-primary-600/40 md:col-span-2 md:row-span-2">
      <Link href={`/campaigns/${item.id}`} className="block">
        <CampaignHeroBanner
          imageUrl={item.image_url}
          label="Spotlight"
          scrim
          className="h-44 sm:h-56 lg:h-60"
        >
          <div className="absolute inset-x-0 bottom-0 p-4 sm:p-5">
            <h2 className="text-[22px] font-extrabold leading-tight tracking-tight text-white sm:text-[26px]">
              {item.title}
            </h2>
            <p className="mt-0.5 text-[13px] font-semibold text-white/80">
              {item.service_name}
              {ending ? ` · ${ending}` : ""}
            </p>
          </div>
        </CampaignHeroBanner>
      </Link>

      <div className="flex flex-1 flex-col gap-4 p-5">
        <DiscountBadge
          bare
          originalPrice={item.first_tier_price}
          currentPrice={item.current_price}
          percentOff={percentOff}
        />

        <JoinedAvatarsRow bare participants={[]} totalJoined={item.participant_count} unlocked={unlocked} />

        <div>
          <ProgressBar value={progress} height={10} />
          <p className="num mt-2 text-[12.5px] font-semibold text-muted">
            {item.participant_count} of {item.min_participants} flats joined
          </p>
        </div>

        <Link
          href={`/campaigns/${item.id}`}
          className="mt-auto inline-flex w-full items-center justify-center gap-1.5 rounded-2xl bg-primary-600 px-5 py-3 text-[15px] font-bold text-white transition-colors hover:bg-primary-700"
        >
          {item.is_joined ? "View my spot" : "Join this deal"}
          <ArrowRight className="h-4 w-4" />
        </Link>
      </div>
    </div>
  );
}
