// Shared read-only content for both the authenticated campaign detail page
// and the public (slug) share page.

import type { CampaignResponse, ParticipantResponse } from "@/lib/types";
import { formatPaise, formatDateLong, formatTime, daysLeft } from "@/lib/format";
import { StatusChip } from "./status-chip";
import { ProgressBar } from "@/components/ui/progress-bar";
import { CampaignHeroBanner, JoinedAvatarsRow, DiscountBadge, ShareButton } from "./campaign-visuals";

export function CampaignDetailBody({
  campaign,
  participants = [],
}: {
  campaign: CampaignResponse;
  participants?: ParticipantResponse[];
}) {
  const progress = campaign.min_participants === 0 ? 0 : campaign.participant_count / campaign.min_participants;
  const minimumReached = campaign.participant_count >= campaign.min_participants;
  const sortedTiers = [...(campaign.pricing_tiers ?? [])].sort((a, b) => a.min_count - b.min_count);
  const ending = daysLeft(campaign.end_date);
  const firstTierPrice = sortedTiers[0]?.price ?? campaign.current_price;
  const percentOff =
    firstTierPrice > campaign.current_price
      ? Math.round(((firstTierPrice - campaign.current_price) / firstTierPrice) * 100)
      : 0;

  return (
    <div>
      <div className="mb-4">
        <CampaignHeroBanner imageUrl={campaign.image_url} />
      </div>

      <div className="mb-1 flex items-start justify-between gap-3">
        <h1 className="text-xl font-bold text-ink">{campaign.title}</h1>
        <StatusChip status={campaign.status} />
      </div>
      <div className="mb-5 flex items-center justify-between">
        <p className="text-sm text-muted">{campaign.service_name}</p>
        <ShareButton title={campaign.title} url={campaign.share_url} />
      </div>

      <div className="mb-5">
        <JoinedAvatarsRow
          participants={participants}
          totalJoined={campaign.participant_count}
          unlocked={minimumReached}
        />
      </div>

      <div className="mb-5">
        <DiscountBadge
          originalPrice={firstTierPrice}
          currentPrice={campaign.current_price}
          percentOff={percentOff}
        />
      </div>

      {/* Progress card */}
      <div className="mb-5 rounded-xl border border-border bg-card p-5">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium text-muted">Flats joined</p>
          <p className="text-sm font-extrabold text-ink">
            {campaign.participant_count} / {campaign.min_participants}
          </p>
        </div>
        <div className="mt-2.5">
          <ProgressBar value={progress} height={10} />
        </div>
        {campaign.next_tier ? (
          <p className="mt-2 text-xs font-semibold text-primary-700 dark:text-primary-400">
            {campaign.participants_to_next_tier} more to unlock {formatPaise(campaign.next_tier.price)}
          </p>
        ) : (
          <p className="mt-2 text-xs font-semibold text-primary-700 dark:text-primary-400">
            Best group price unlocked 🎉
          </p>
        )}
      </div>

      {campaign.description && (
        <div className="mb-5">
          <h2 className="mb-1.5 text-sm font-semibold text-ink">About this deal</h2>
          <p className="text-sm leading-relaxed text-muted">{campaign.description}</p>
        </div>
      )}

      {sortedTiers.length > 0 && (
        <div className="mb-5">
          <h2 className="mb-2 text-sm font-semibold text-ink">Pricing tiers</h2>
          <div className="flex flex-col gap-2">
            {sortedTiers.map((tier) => {
              const active =
                campaign.participant_count >= tier.min_count &&
                (tier.max_count === 0 || campaign.participant_count <= tier.max_count);
              return (
                <div
                  key={tier.min_count}
                  className={`flex items-center justify-between rounded-lg border px-3.5 py-2.5 text-sm ${
                    active
                      ? "border-primary-600/40 bg-primary-50 dark:bg-primary-900/20"
                      : "border-border"
                  }`}
                >
                  <span className="font-medium text-ink">
                    {tier.max_count === 0 ? `${tier.min_count}+ people` : `${tier.min_count}–${tier.max_count} people`}
                  </span>
                  <span className={`font-bold ${active ? "text-primary-700 dark:text-primary-400" : "text-ink"}`}>
                    {formatPaise(tier.price)}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}

      <div className="rounded-xl border border-border bg-card p-4 text-sm">
        <InfoRow label="Service date" value={formatDateLong(campaign.service_date)} />
        <InfoRow label="Time" value={formatTime(campaign.service_date)} />
        <InfoRow label="Deal closes" value={formatDateLong(campaign.end_date)} />
        {ending && <InfoRow label="Days remaining" value={ending} last />}
      </div>
    </div>
  );
}

function InfoRow({ label, value, last = false }: { label: string; value: string; last?: boolean }) {
  return (
    <div className={`flex items-center justify-between py-2 ${last ? "" : "border-b border-border"}`}>
      <span className="text-muted">{label}</span>
      <span className="font-medium text-ink">{value}</span>
    </div>
  );
}
