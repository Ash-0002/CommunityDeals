// Editorial campaign layout, shared by the authenticated detail page
// (/campaigns/[id]) and the public share page (/c/[slug]).
//
//   ▸ full-bleed hero image, title + status overlaid on a gradient scrim
//   ▸ two-column body at lg and up:
//       left  — description, pricing tiers, booking details
//       right — sticky join card (price, savings, avatars, progress, CTA,
//               service date & time, location)
//   ▸ single column below lg; the join card floats up to just under the hero
//     so a phone visitor sees the price and CTA without scrolling far.

import { CalendarDays, Clock, MapPin, Users } from "lucide-react";
import type { CampaignResponse, CommunityResponse, ParticipantResponse } from "@/lib/types";
import {
  formatPaise,
  formatDateLong,
  formatTime,
  daysLeft,
  shortLocation,
  fullLocation,
} from "@/lib/format";
import { StatusChip } from "./status-chip";
import { ProgressBar } from "@/components/ui/progress-bar";
import { CampaignHeroBanner, JoinedAvatarsRow, DiscountBadge, ShareButton } from "./campaign-visuals";

export function CampaignEditorial({
  campaign,
  participants = [],
  community = null,
  /** Join / cancel control rendered inside the sticky card. */
  action,
  /** Small reassurance line under the CTA. */
  actionNote,
  /** Extra blocks rendered under the left column (e.g. cancel confirmation). */
  children,
}: {
  campaign: CampaignResponse;
  participants?: ParticipantResponse[];
  community?: CommunityResponse | null;
  action?: React.ReactNode;
  actionNote?: React.ReactNode;
  children?: React.ReactNode;
}) {
  const progress =
    campaign.min_participants === 0 ? 0 : campaign.participant_count / campaign.min_participants;
  const minimumReached = campaign.participant_count >= campaign.min_participants;
  const sortedTiers = [...(campaign.pricing_tiers ?? [])].sort((a, b) => a.min_count - b.min_count);
  const ending = daysLeft(campaign.end_date);
  const firstTierPrice = sortedTiers[0]?.price ?? campaign.current_price;
  const percentOff =
    firstTierPrice > campaign.current_price
      ? Math.round(((firstTierPrice - campaign.current_price) / firstTierPrice) * 100)
      : 0;

  const place = fullLocation(community) || shortLocation(community);

  return (
    <article>
      {/* ── Hero ─────────────────────────────────────────────────────────── */}
      <CampaignHeroBanner
        imageUrl={campaign.image_url}
        label={null}
        scrim
        className="h-[260px] rounded-3xl sm:h-[340px] lg:h-[420px]"
      >
        <div className="absolute inset-x-0 top-0 flex items-start justify-between gap-3 p-4 sm:p-6">
          <span className="inline-flex min-w-0 items-center gap-1.5 rounded-full bg-white/95 px-3 py-1.5 text-[11px] font-extrabold uppercase tracking-wide text-primary-700 shadow-sm">
            <Users className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{campaign.service_name}</span>
          </span>
          <ShareButton title={campaign.title} url={campaign.share_url} variant="overlay" />
        </div>

        <div className="absolute inset-x-0 bottom-0 p-4 sm:p-6 lg:p-8">
          <div className="mb-3 flex flex-wrap items-center gap-2">
            <StatusChip status={campaign.status} />
            {ending && (
              <span className="inline-flex items-center rounded-full bg-black/40 px-2.5 py-0.5 text-[11px] font-semibold text-white">
                {ending}
              </span>
            )}
          </div>
          <h1 className="max-w-3xl text-[26px] font-extrabold leading-[1.1] tracking-tight text-white sm:text-[38px] lg:text-[46px]">
            {campaign.title}
          </h1>
          {community && (
            <p className="mt-2 flex items-center gap-1.5 text-[13px] font-semibold text-white/85 sm:text-sm">
              <MapPin className="h-4 w-4 shrink-0" />
              <span className="truncate">
                {community.name}
                {shortLocation(community) ? ` · ${shortLocation(community)}` : ""}
              </span>
            </p>
          )}
        </div>
      </CampaignHeroBanner>

      {/* ── Body ─────────────────────────────────────────────────────────── */}
      <div className="mt-6 grid items-start gap-6 lg:mt-8 lg:grid-cols-[minmax(0,1fr)_360px] lg:gap-8">
        {/* Join card. First in source order so it lands directly under the hero
            on phones; moves to the right rail and sticks at lg. */}
        <aside className="order-first lg:order-last lg:sticky lg:top-28">
          <div className="rounded-3xl border border-border bg-card p-5 shadow-card sm:p-6">
            <DiscountBadge
              bare
              originalPrice={firstTierPrice}
              currentPrice={campaign.current_price}
              percentOff={percentOff}
            />

            <div className="mt-5 border-t border-border pt-5">
              <JoinedAvatarsRow
                bare
                participants={participants}
                totalJoined={campaign.participant_count}
                unlocked={minimumReached}
              />
            </div>

            <div className="mt-5">
              <div className="mb-2 flex items-baseline justify-between">
                <span className="text-xs font-semibold text-muted">Flats joined</span>
                <span className="num text-sm font-extrabold text-ink">
                  {campaign.participant_count} / {campaign.min_participants}
                </span>
              </div>
              <ProgressBar value={progress} height={10} />
              <p className="mt-2 text-xs font-semibold text-primary-700 dark:text-primary-400">
                {campaign.next_tier
                  ? `${campaign.participants_to_next_tier} more to unlock ${formatPaise(
                      campaign.next_tier.price,
                    )}`
                  : "Best group price unlocked 🎉"}
              </p>
            </div>

            {action && <div className="mt-5">{action}</div>}
            {actionNote && (
              <div className="mt-3 text-center text-[11.5px] leading-relaxed text-muted">
                {actionNote}
              </div>
            )}

            <div className="mt-5 space-y-3 border-t border-border pt-5">
              <IconRow
                icon={CalendarDays}
                label="Service date"
                value={formatDateLong(campaign.service_date)}
              />
              <IconRow icon={Clock} label="Time" value={formatTime(campaign.service_date)} />
              {place && <IconRow icon={MapPin} label="Location" value={place} />}
            </div>
          </div>
        </aside>

        {/* Left column — the read */}
        <div className="min-w-0 space-y-6">
          {campaign.description && (
            <section className="rounded-3xl border border-border bg-card p-5 sm:p-6">
              <SectionTitle>About this deal</SectionTitle>
              <p className="text-[15px] leading-relaxed text-muted">{campaign.description}</p>
            </section>
          )}

          {sortedTiers.length > 0 && (
            <section className="rounded-3xl border border-border bg-card p-5 sm:p-6">
              <SectionTitle>How the price drops</SectionTitle>
              <div className="flex flex-col gap-2">
                {sortedTiers.map((tier) => {
                  const active =
                    campaign.participant_count >= tier.min_count &&
                    (tier.max_count === 0 || campaign.participant_count <= tier.max_count);
                  return (
                    <div
                      key={tier.min_count}
                      className={`flex items-center justify-between gap-3 rounded-2xl border px-4 py-3 text-sm transition-colors ${
                        active
                          ? "border-primary-600/40 bg-primary-50 dark:bg-primary-900/20"
                          : "border-border"
                      }`}
                    >
                      <span className="flex min-w-0 items-center gap-2 font-semibold text-ink">
                        <Users className="h-4 w-4 shrink-0 text-muted" />
                        <span className="truncate">
                          {tier.max_count === 0
                            ? `${tier.min_count}+ people`
                            : `${tier.min_count}–${tier.max_count} people`}
                        </span>
                      </span>
                      <span
                        className={`num shrink-0 text-[15px] font-extrabold ${
                          active ? "text-primary-700 dark:text-primary-400" : "text-ink"
                        }`}
                      >
                        {formatPaise(tier.price)}
                      </span>
                    </div>
                  );
                })}
              </div>
            </section>
          )}

          <section className="rounded-3xl border border-border bg-card p-5 sm:p-6">
            <SectionTitle>Booking details</SectionTitle>
            <div className="text-sm">
              <InfoRow label="Service" value={campaign.service_name} />
              <InfoRow label="Service date" value={formatDateLong(campaign.service_date)} />
              <InfoRow label="Time" value={formatTime(campaign.service_date)} />
              <InfoRow label="Joining closes" value={formatDateLong(campaign.end_date)} />
              {community && <InfoRow label="Community" value={community.name} />}
              {place && <InfoRow label="Service location" value={place} icon />}
              {ending && <InfoRow label="Days remaining" value={ending} last />}
            </div>
          </section>

          {children}
        </div>
      </div>
    </article>
  );
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="mb-3 text-[11px] font-extrabold uppercase tracking-[0.12em] text-muted">
      {children}
    </h2>
  );
}

function IconRow({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-start gap-3">
      <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
        <Icon className="h-3.5 w-3.5" />
      </span>
      <div className="min-w-0">
        <p className="text-[11px] font-semibold uppercase tracking-wide text-muted">{label}</p>
        <p className="text-[13.5px] font-semibold leading-snug text-ink">{value}</p>
      </div>
    </div>
  );
}

function InfoRow({
  label,
  value,
  last = false,
  icon = false,
}: {
  label: string;
  value: string;
  last?: boolean;
  icon?: boolean;
}) {
  return (
    <div
      className={`flex flex-wrap items-start justify-between gap-x-4 gap-y-0.5 py-2.5 ${
        last ? "" : "border-b border-border"
      }`}
    >
      <span className="shrink-0 text-muted">{label}</span>
      <span className="flex min-w-0 items-center gap-1.5 text-right font-semibold text-ink">
        {icon && <MapPin className="h-3.5 w-3.5 shrink-0 text-primary-600" />}
        {value}
      </span>
    </div>
  );
}
