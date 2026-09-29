import Link from "next/link";
import { Check, Users } from "lucide-react";
import type { CampaignListItem } from "@/lib/types";
import { formatPaise, daysLeft } from "@/lib/format";
import { StatusChip } from "./status-chip";
import { ProgressBar } from "@/components/ui/progress-bar";

export function CampaignCard({ campaign }: { campaign: CampaignListItem }) {
  const progress =
    campaign.min_participants === 0 ? 0 : campaign.participant_count / campaign.min_participants;
  const ending = daysLeft(campaign.end_date);
  const unlocked =
    campaign.min_participants > 0 && campaign.participant_count >= campaign.min_participants;
  const percentOff =
    campaign.first_tier_price > campaign.current_price
      ? Math.round(
          ((campaign.first_tier_price - campaign.current_price) / campaign.first_tier_price) * 100,
        )
      : 0;

  return (
    <Link href={`/campaigns/${campaign.id}`} className="group block h-full">
      <article className="flex h-full flex-col overflow-hidden rounded-3xl border border-border bg-card shadow-card transition-all hover:-translate-y-0.5 hover:border-primary-600/40 hover:shadow-card-hover">
        <div className="relative h-36 w-full shrink-0 bg-gradient-to-br from-primary-100 to-amber-50 dark:from-primary-900/40 dark:to-amber-950/20 sm:h-40">
          {campaign.image_url && (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={campaign.image_url}
              alt=""
              className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
            />
          )}
          <span className="absolute right-2.5 top-2.5">
            <StatusChip status={campaign.status} />
          </span>
          {unlocked && (
            <span className="absolute left-2.5 top-2.5 inline-flex items-center gap-1 rounded-full bg-white/95 px-2.5 py-1 text-[10.5px] font-extrabold text-primary-700 shadow-sm">
              <Check className="h-3 w-3" strokeWidth={3} />
              Unlocked
            </span>
          )}
          {percentOff > 0 && (
            <span className="absolute bottom-2.5 left-2.5 inline-flex items-center rounded-full bg-primary-600 px-2.5 py-1 text-[11px] font-extrabold text-white shadow-sm">
              {percentOff}% off
            </span>
          )}
        </div>

        <div className="flex flex-1 flex-col p-4 sm:p-5">
          <h3 className="text-[16px] font-extrabold leading-snug tracking-tight text-ink">
            {campaign.title}
          </h3>
          <p className="mt-0.5 text-[12.5px] font-medium text-muted">{campaign.service_name}</p>

          <div className="mt-4">
            <ProgressBar value={progress} />
          </div>

          <div className="mt-2.5 flex items-center justify-between gap-2">
            <span className="num flex items-center gap-1.5 text-[12.5px] font-semibold text-muted">
              <Users className="h-3.5 w-3.5" />
              {campaign.participant_count}/{campaign.min_participants} joined
            </span>
            <span className="num text-[17px] font-black text-primary-700 dark:text-primary-400">
              {formatPaise(campaign.current_price)}
            </span>
          </div>

          {ending && (
            <p className="mt-auto pt-3 text-[11.5px] font-semibold text-muted">{ending}</p>
          )}
        </div>
      </article>
    </Link>
  );
}
