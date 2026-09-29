"use client";

import { Users, Check, Share2 } from "lucide-react";
import { formatPaise } from "@/lib/format";
import type { ParticipantResponse } from "@/lib/types";

// ── Hero image banner ───────────────────────────────────────────────────────

export function CampaignHeroBanner({
  imageUrl,
  label = "Group booking",
}: {
  imageUrl?: string;
  label?: string;
}) {
  return (
    <div className="relative aspect-[16/9] w-full overflow-hidden rounded-2xl bg-gradient-to-br from-blue-50 to-sky-50 dark:from-blue-950/40 dark:to-sky-950/30">
      {imageUrl && (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={imageUrl} alt="" className="h-full w-full object-cover" />
      )}
      <span className="absolute left-3 top-3 inline-flex items-center gap-1.5 rounded-full bg-white/95 px-3 py-1.5 text-[11px] font-extrabold uppercase tracking-wide text-primary-700 shadow-sm">
        <Users className="h-3.5 w-3.5" />
        {label}
      </span>
    </div>
  );
}

// ── Avatar stack + unlocked banner ──────────────────────────────────────────

export function JoinedAvatarsRow({
  participants,
  totalJoined,
  unlocked,
}: {
  participants: ParticipantResponse[];
  totalJoined: number;
  unlocked: boolean;
}) {
  const maxShown = 6;
  const shown = participants.slice(0, maxShown);
  const overflow = totalJoined - shown.length;

  return (
    <div className="rounded-2xl border border-border bg-card p-4">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-primary-600">
          <Users className="h-5 w-5 text-white" />
        </div>
        <span className="rounded-md bg-amber-200/80 px-2.5 py-1 text-[13px] dark:bg-amber-400/20">
          <span className="font-extrabold text-ink">{totalJoined} </span>
          <span className="font-semibold text-ink">people joined</span>
        </span>
      </div>

      {shown.length > 0 && (
        <div className="relative mt-4 h-10">
          {shown.map((p, i) => (
            <div
              key={p.user_id}
              className="absolute top-0 flex h-9 w-9 items-center justify-center overflow-hidden rounded-full border-2 border-card bg-primary-100 text-[11px] font-extrabold text-primary-700 dark:bg-primary-900/50 dark:text-primary-300"
              style={{ left: i * 26, zIndex: shown.length - i }}
              title={p.name}
            >
              {p.avatar_url ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={p.avatar_url} alt={p.name} className="h-full w-full object-cover" />
              ) : (
                initials(p.name)
              )}
            </div>
          ))}
          {overflow > 0 && (
            <div
              className="absolute top-0 flex h-9 w-9 items-center justify-center rounded-full border-2 border-card bg-primary-50 text-[11px] font-extrabold text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
              style={{ left: shown.length * 26 }}
            >
              +{overflow}
            </div>
          )}
        </div>
      )}

      {unlocked && (
        <div className="mt-4 flex justify-start">
          <span className="unlock-pop inline-flex items-center gap-2 rounded-full bg-primary-50 px-3.5 py-1.5 dark:bg-primary-900/30">
            <span className="flex h-[18px] w-[18px] items-center justify-center rounded-full bg-primary-600">
              <Check className="h-3 w-3 text-white" strokeWidth={3} />
            </span>
            <span className="text-[12.5px] font-extrabold text-primary-700 dark:text-primary-300">
              Group booking unlocked
            </span>
          </span>
        </div>
      )}
    </div>
  );
}

function initials(name: string): string {
  const trimmed = name.trim();
  if (!trimmed) return "?";
  const parts = trimmed.split(/\s+/);
  if (parts.length === 1) return parts[0]![0]!.toUpperCase();
  return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
}

// ── Discount badge ──────────────────────────────────────────────────────────

export function DiscountBadge({
  originalPrice,
  currentPrice,
  percentOff,
}: {
  originalPrice: number;
  currentPrice: number;
  percentOff: number;
}) {
  return (
    <div className="flex items-center gap-4 rounded-2xl border border-border bg-card p-4">
      {percentOff > 0 && (
        <div className="starburst flex h-[72px] w-[72px] flex-shrink-0 items-center justify-center text-white">
          <div className="flex flex-col items-center leading-none">
            <span className="text-[17px] font-black">{percentOff}%</span>
            <span className="text-[10px] font-extrabold tracking-wide">OFF</span>
          </div>
        </div>
      )}
      <div className="min-w-0 flex-1">
        <p className="text-xs text-muted">Exclusive group price</p>
        <div className="mt-1 flex items-end gap-2">
          {originalPrice > currentPrice && (
            <span className="text-sm text-muted line-through">{formatPaise(originalPrice)}</span>
          )}
          <span className="text-2xl font-black text-primary-700 dark:text-primary-400">
            {formatPaise(currentPrice)}
          </span>
        </div>
        {originalPrice > currentPrice && (
          <span className="mt-1 inline-block rounded-full bg-primary-50 px-2 py-0.5 text-[11px] font-bold text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
            You save {formatPaise(originalPrice - currentPrice)}
          </span>
        )}
      </div>
    </div>
  );
}

// ── Share button ─────────────────────────────────────────────────────────────

export function ShareButton({ title, url }: { title: string; url: string }) {
  async function handleShare() {
    const text = `Join "${title}" on CommunityDeals — ${url}`;
    if (typeof navigator !== "undefined" && navigator.share) {
      try {
        await navigator.share({ title, text, url });
        return;
      } catch {
        // user cancelled or share failed — fall through to clipboard
      }
    }
    try {
      await navigator.clipboard.writeText(text);
      alert("Link copied to clipboard");
    } catch {
      // ignore
    }
  }

  return (
    <button
      onClick={handleShare}
      className="flex h-8 w-8 items-center justify-center rounded-full text-primary-700 hover:bg-primary-50 dark:text-primary-400 dark:hover:bg-primary-900/30"
      title="Share"
    >
      <Share2 className="h-4 w-4" />
    </button>
  );
}
