import type { CampaignStatus } from "@/lib/types";

const STYLES: Record<CampaignStatus, { label: string; cls: string }> = {
  DRAFT: { label: "Draft", cls: "bg-surface text-muted" },
  PUBLISHED: { label: "Open", cls: "bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300" },
  MINIMUM_REACHED: { label: "Min reached", cls: "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300" },
  CONFIRMED: { label: "Confirmed", cls: "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300" },
  COMPLETED: { label: "Completed", cls: "bg-surface text-muted" },
  CANCELLED: { label: "Cancelled", cls: "bg-red-50 text-red-600 dark:bg-red-950/40 dark:text-red-400" },
};

export function StatusChip({ status }: { status: CampaignStatus }) {
  const s = STYLES[status] ?? STYLES.DRAFT;
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-semibold ${s.cls}`}>
      {s.label}
    </span>
  );
}
