"use client";

import Link from "next/link";
import type { CampaignResponse } from "@/lib/types";

function formatPrice(paise: number) {
  return `₹${(paise / 100).toLocaleString("en-IN")}`;
}

function formatDate(iso: string | null) {
  if (!iso) return "Date TBD";
  return new Date(iso).toLocaleDateString("en-IN", {
    weekday: "long", day: "numeric", month: "long",
  });
}

function formatTime(iso: string | null) {
  if (!iso) return "Time TBD";
  return new Date(iso).toLocaleTimeString("en-IN", {
    hour: "2-digit", minute: "2-digit", hour12: true,
  });
}

export function CampaignPublicView({ campaign }: { campaign: CampaignResponse }) {
  const fill = Math.min(100, Math.round((campaign.participant_count / campaign.min_participants) * 100));
  const needed = Math.max(0, campaign.min_participants - campaign.participant_count);
  const isLocked = campaign.status === "CONFIRMED" || campaign.status === "MINIMUM_REACHED";

  // Sort tiers by min_count ascending
  const sortedTiers = [...(campaign.pricing_tiers ?? [])].sort((a, b) => a.min_count - b.min_count);

  return (
    <div className="min-h-screen bg-surface">
      {/* Top nav — visible to unauthenticated visitors */}
      <header className="border-b border-border bg-card px-6 py-3.5">
        <div className="mx-auto flex max-w-5xl items-center justify-between">
          <Link href="/" className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary-600 text-xs font-black text-white">
              C
            </div>
            <span className="text-sm font-extrabold text-ink">CommunityDeals</span>
          </Link>
          <Link
            href="/auth/login"
            className="rounded-lg bg-primary-600 px-4 py-1.5 text-xs font-bold text-white hover:bg-primary-700"
          >
            Log in
          </Link>
        </div>
      </header>

      <div className="mx-auto max-w-5xl px-6 py-8">
        <div className="grid grid-cols-[1fr_320px] gap-6 items-start">

          {/* ── Left column ── */}
          <div>
            {/* Hero */}
            <div className="relative mb-5 overflow-hidden rounded-2xl bg-gradient-to-br from-sky-700 to-sky-900 p-7">
              <div className="pointer-events-none absolute right-[-20px] top-[-20px] h-24 w-44 rotate-12 rounded-2xl border border-white/10 bg-white/5" />
              <div className="pointer-events-none absolute bottom-[-30px] right-10 h-28 w-28 rounded-full bg-white/[0.04]" />

              <div className="mb-3 w-fit rounded-full border border-white/20 bg-white/15 px-3 py-1 text-[10px] font-bold uppercase tracking-widest text-white backdrop-blur-sm">
                Group Booking
              </div>
              <h1 className="mb-1.5 text-3xl font-extrabold leading-tight text-white">
                {campaign.title}
              </h1>
              <p className="flex items-center gap-1.5 text-[13px] font-medium text-white/60">
                📍 {campaign.community_name}
                {campaign.service_date && <> · {formatDate(campaign.service_date)}</>}
              </p>

              <div className="mt-5 flex gap-6 border-t border-white/10 pt-5">
                {[
                  { val: campaign.participant_count, label: "Flats joined" },
                  { val: campaign.min_participants,  label: "Target"       },
                  { val: needed,                     label: "More needed"  },
                ].map((s) => (
                  <div key={s.label}>
                    <p className="num text-xl font-extrabold text-white">{s.val}</p>
                    <p className="text-[11px] font-medium text-white/50">{s.label}</p>
                  </div>
                ))}
              </div>
            </div>

            {/* Description */}
            {campaign.description && (
              <div className="mb-4 rounded-xl bg-card border border-border p-5 shadow-card">
                <h2 className="mb-2 text-[13px] font-bold text-ink">About this campaign</h2>
                <p className="text-sm leading-relaxed text-muted">{campaign.description}</p>
              </div>
            )}

            {/* Pricing tiers */}
            {sortedTiers.length > 0 && (
              <div className="mb-4 rounded-xl bg-card border border-border p-5 shadow-card">
                <h2 className="mb-3.5 text-[13px] font-bold text-ink">
                  Pricing Tiers — group discount unlocks at {campaign.min_participants} flats
                </h2>
                <table className="w-full">
                  <thead>
                    <tr className="text-left">
                      <th className="pb-2 text-[10px] font-bold uppercase tracking-wider text-muted">Flats</th>
                      <th className="pb-2 text-[10px] font-bold uppercase tracking-wider text-muted">Price per unit</th>
                      <th className="pb-2 text-[10px] font-bold uppercase tracking-wider text-muted">You save</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    {sortedTiers.map((tier, i) => {
                      const isActive = campaign.current_price === tier.price;
                      const basePrice = sortedTiers[0]?.price ?? tier.price;
                      const saved = basePrice - tier.price;
                      return (
                        <tr
                          key={tier.min_count}
                          className={isActive ? "bg-primary-600/10 dark:bg-primary-600/20" : ""}
                        >
                          <td className={`border-t border-border py-2.5 text-[13px] font-semibold ${isActive ? "text-primary-600 pl-2 rounded-l-lg" : "text-muted"}`}>
                            {tier.min_count}+ flats
                          </td>
                          <td className="num border-t border-border py-2.5 text-[13px] font-semibold text-ink">
                            {formatPrice(tier.price)}
                          </td>
                          <td className="num border-t border-border py-2.5 text-[13px] font-semibold text-primary-500">
                            {i === 0 ? "—" : `${formatPrice(saved)} off`}
                          </td>
                          <td className={`border-t border-border py-2.5 ${isActive ? "pr-2 rounded-r-lg" : ""}`}>
                            {isActive && (
                              <span className="rounded-full bg-primary-600/15 px-2.5 py-0.5 text-[9px] font-bold uppercase tracking-wide text-primary-500">
                                ● Current
                              </span>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}

            {/* Service details */}
            <div className="mb-4 rounded-xl bg-card border border-border p-5 shadow-card">
              <h2 className="mb-3.5 text-[13px] font-bold text-ink">Service Details</h2>
              <div className="grid grid-cols-2 gap-3">
                {[
                  { icon: "📅", label: "Service Date", val: formatDate(campaign.service_date) },
                  { icon: "⏰", label: "Time",         val: formatTime(campaign.service_date) },
                  { icon: "🏢", label: "Vendor",       val: campaign.vendor_name || "TBD"    },
                  { icon: "📍", label: "Location",     val: campaign.location_name || campaign.city || "TBD" },
                ].map((item) => (
                  <div key={item.label} className="flex items-center gap-3">
                    <div className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-surface text-base">
                      {item.icon}
                    </div>
                    <div>
                      <p className="text-[10px] font-semibold uppercase tracking-wide text-muted">{item.label}</p>
                      <p className="text-[13px] font-bold text-ink">{item.val}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* ── Right: sticky join panel ── */}
          <div className="sticky top-6">
            <div className="rounded-2xl bg-card border border-border p-5 shadow-card">
              {/* Counter */}
              <div className="mb-4 text-center">
                <span className="num text-5xl font-extrabold text-ink">
                  {campaign.participant_count}
                </span>
                <span className="num text-2xl font-semibold text-muted">
                  /{campaign.min_participants}
                </span>
                <p className="mt-1 text-xs font-semibold text-muted">flats joined</p>
              </div>

              {/* Progress bar */}
              <div className="mb-2 h-2 overflow-hidden rounded-full bg-border">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-primary-600 to-primary-400 transition-all"
                  style={{ width: `${fill}%` }}
                />
              </div>

              {/* Urgency strip */}
              {!isLocked && needed > 0 && (
                <p className="mb-4 rounded-lg bg-amber-500/10 px-3 py-2 text-center text-[11px] font-semibold text-amber-600 dark:text-amber-400">
                  ⚡ {needed} more {needed === 1 ? "flat" : "flats"} needed to unlock group pricing
                </p>
              )}
              {isLocked && (
                <p className="mb-4 rounded-lg bg-primary-600/10 px-3 py-2 text-center text-[11px] font-semibold text-primary-500">
                  ✅ Group booking unlocked!
                </p>
              )}

              {/* Current price box */}
              <div className="mb-4 rounded-xl border border-primary-600/20 bg-primary-600/10 p-4 text-center">
                <p className="text-[10px] font-bold uppercase tracking-wider text-muted">
                  {isLocked ? "Locked Price" : "Current Price"}
                </p>
                <p className="num mt-1 text-3xl font-extrabold text-primary-500">
                  {formatPrice(campaign.current_price)}
                </p>
                <p className="mt-0.5 text-[11px] font-semibold text-muted">per unit</p>
              </div>

              {/* Join CTA */}
              <Link
                href={`/auth/login?redirect=/c/${campaign.slug}`}
                className="mb-2.5 flex w-full items-center justify-center gap-2 rounded-xl bg-primary-600 py-3.5 text-[15px] font-bold text-white shadow-[0_4px_14px_rgba(22,163,74,0.3)] hover:bg-primary-700 transition-colors"
              >
                👥 Join Group — {formatPrice(campaign.current_price)}
              </Link>

              <button className="flex w-full items-center justify-center gap-1.5 rounded-xl border border-border py-2.5 text-[13px] font-semibold text-muted hover:border-primary-600/40 hover:text-primary-500 transition-colors">
                📤 Share on WhatsApp
              </button>

              <p className="mt-3 text-center text-[10px] text-muted">
                No payment now · Pay only when the group is confirmed
              </p>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}
