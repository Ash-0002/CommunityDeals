import Link from "next/link";
import type { CampaignResponse } from "@/lib/types";

const GRADIENTS: Record<string, string> = {
  AC_SERVICE:   "from-sky-700 to-sky-900",
  PEST_CONTROL: "from-red-700 to-red-900",
  WATER_TANK:   "from-violet-700 to-violet-900",
  CAR_WASH:     "from-amber-600 to-orange-900",
  PLUMBING:     "from-slate-600 to-slate-900",
  ELECTRICAL:   "from-yellow-600 to-yellow-900",
};
const DEFAULT_GRADIENT = "from-slate-600 to-slate-900";

function statusChip(c: CampaignResponse) {
  const needed = c.min_participants - c.participant_count;
  if (c.status === "CONFIRMED" || c.status === "MINIMUM_REACHED")
    return { label: "Group locked ✓", cls: "bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300" };
  if (needed <= 3 && needed > 0)
    return { label: `${needed} more needed`, cls: "bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300" };
  return { label: "Open", cls: "bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-300" };
}

function formatPrice(paise: number) {
  return `₹${(paise / 100).toLocaleString("en-IN")}`;
}

export function CampaignCard({ campaign }: { campaign: CampaignResponse }) {
  const gradient = GRADIENTS[campaign.service_type] ?? DEFAULT_GRADIENT;
  const fill = Math.min(100, Math.round((campaign.participant_count / campaign.min_participants) * 100));
  const chip = statusChip(campaign);
  const date = campaign.service_date
    ? new Date(campaign.service_date).toLocaleDateString("en-IN", { day: "numeric", month: "short" })
    : null;

  return (
    <Link href={`/c/${campaign.slug}`} className="group block">
      <div className="overflow-hidden rounded-2xl bg-card shadow-card transition-shadow group-hover:shadow-card-hover">
        {/* Coloured header */}
        <div className={`relative bg-gradient-to-br ${gradient} p-3 pt-8`}>
          <div className="absolute left-3 top-3 rounded-full border border-white/20 bg-white/15 px-2.5 py-0.5 text-[9px] font-bold uppercase tracking-wide text-white backdrop-blur-sm">
            {campaign.service_type.replace(/_/g, " ")}
          </div>
          <p className="text-lg font-extrabold leading-snug text-white">{campaign.title}</p>
          {campaign.community_name && (
            <p className="mt-0.5 text-[11px] font-medium text-white/60">{campaign.community_name}</p>
          )}
        </div>

        {/* Body */}
        <div className="p-3.5">
          <div className="mb-2 flex items-baseline justify-between">
            <span className="num text-xl font-extrabold text-ink">
              {campaign.participant_count}
              <span className="text-sm font-semibold text-muted"> / {campaign.min_participants}</span>
            </span>
            <span className="num text-base font-extrabold text-primary-600">{formatPrice(campaign.current_price)}</span>
          </div>

          <div className="mb-3 h-1 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-700">
            <div
              className="h-full rounded-full bg-gradient-to-r from-primary-600 to-primary-400 transition-all"
              style={{ width: `${fill}%` }}
            />
          </div>

          <div className="flex items-center justify-between">
            <span className="text-[10px] font-medium text-muted">
              {date ? `📍 ${date}` : "📍 Date TBD"}
            </span>
            <span className={`rounded-full px-2.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${chip.cls}`}>
              {chip.label}
            </span>
          </div>
        </div>
      </div>
    </Link>
  );
}
