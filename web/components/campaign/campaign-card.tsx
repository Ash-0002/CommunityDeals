import Link from "next/link";
import { Check } from "lucide-react";
import type { CampaignListItem } from "@/lib/types";
import { formatPaise, daysLeft } from "@/lib/format";
import { StatusChip } from "./status-chip";
import { ProgressBar } from "@/components/ui/progress-bar";

export function CampaignCard({ campaign }: { campaign: CampaignListItem }) {
  const progress = campaign.min_participants === 0 ? 0 : campaign.participant_count / campaign.min_participants;
  const ending = daysLeft(campaign.end_date);
  const unlocked = campaign.min_participants > 0 && campaign.participant_count >= campaign.min_participants;

  return (
    <Link href={`/campaigns/${campaign.id}`} className="block">
      <div className="overflow-hidden rounded-xl border border-border bg-card transition-colors hover:border-primary-600/40">
        {campaign.image_url && (
          <div className="relative h-28 w-full">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src={campaign.image_url} alt="" className="h-full w-full object-cover" />
            <span className="absolute right-2 top-2">
              <StatusChip status={campaign.status} />
            </span>
            {unlocked && (
              <span className="absolute left-2 top-2 inline-flex items-center gap-1 rounded-full bg-white/95 px-2 py-1 text-[10.5px] font-extrabold text-primary-700">
                <Check className="h-3 w-3" strokeWidth={3} />
                Unlocked
              </span>
            )}
          </div>
        )}
        <div className="p-4">
          <div className="mb-2 flex items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="truncate text-[14px] font-semibold text-ink">{campaign.title}</p>
              <p className="text-[12px] text-muted">{campaign.service_name}</p>
            </div>
            {!campaign.image_url && <StatusChip status={campaign.status} />}
          </div>

          <div className="mb-2 mt-3">
            <ProgressBar value={progress} />
          </div>

          <div className="flex items-center justify-between">
            <span className="text-[12px] font-medium text-muted">
              {campaign.participant_count}/{campaign.min_participants} joined
            </span>
            <span className="text-[14px] font-bold text-primary-600">{formatPaise(campaign.current_price)}</span>
          </div>
          {ending && <p className="mt-1.5 text-[11px] text-muted">{ending}</p>}
        </div>
      </div>
    </Link>
  );
}
