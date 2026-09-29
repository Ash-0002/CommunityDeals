"use client";

// Deals feed — /campaigns
// Responsive card grid (1 / 2 / 3 columns) with a horizontally scrollable
// filter rail so the chips never wrap awkwardly on a phone.

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { CampaignCard } from "@/components/campaign/campaign-card";
import { api, APIError } from "@/lib/api";
import type { CampaignListItem, CampaignStatus } from "@/lib/types";

const FILTERS: { label: string; status?: CampaignStatus }[] = [
  { label: "All" },
  { label: "Open", status: "PUBLISHED" },
  { label: "Min reached", status: "MINIMUM_REACHED" },
  { label: "Confirmed", status: "CONFIRMED" },
];

export default function CampaignsPage() {
  const [active, setActive] = useState(FILTERS[0]!);
  const [campaigns, setCampaigns] = useState<CampaignListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");
    api.campaigns
      .list({ status: active.status, limit: 50 })
      .then((page) => {
        if (!cancelled) setCampaigns(page.campaigns);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof APIError ? err.message : "Failed to load campaigns");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [active]);

  return (
    <AppLayout>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[13px] font-semibold uppercase tracking-[0.14em] text-muted">
            Open in your society
          </p>
          <h1 className="mt-1.5 text-[30px] font-extrabold leading-tight tracking-tight text-ink sm:text-[40px]">
            Deals
          </h1>
        </div>
        <Link
          href="/campaigns/new"
          className="hidden items-center gap-1.5 rounded-full bg-primary-600 px-5 py-2.5 text-[14px] font-bold text-white transition-colors hover:bg-primary-700 sm:inline-flex"
        >
          <Plus className="h-4 w-4" strokeWidth={3} />
          Start a deal
        </Link>
      </div>

      {/* Filter rail — scrolls sideways instead of wrapping on narrow screens */}
      <div className="-mx-4 mb-6 overflow-x-auto px-4 pb-1 sm:mx-0 sm:px-0">
        <div className="flex w-max gap-2">
          {FILTERS.map((f) => (
            <button
              key={f.label}
              onClick={() => setActive(f)}
              aria-pressed={active.label === f.label}
              className={`shrink-0 rounded-full px-4 py-2 text-[13px] font-bold transition-colors ${
                active.label === f.label
                  ? "bg-primary-600 text-white"
                  : "border border-border bg-card text-muted hover:text-ink"
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <p className="text-sm text-muted">Loading…</p>
      ) : error ? (
        <p className="text-sm text-red-500">{error}</p>
      ) : campaigns.length === 0 ? (
        <div className="rounded-3xl border border-dashed border-border p-10 text-center">
          <p className="text-sm font-semibold text-ink">Nothing here yet.</p>
          <p className="mt-1 text-sm text-muted">
            Be the first —{" "}
            <Link href="/campaigns/new" className="font-semibold text-primary-600">
              start a deal
            </Link>
            .
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {campaigns.map((c) => (
            <CampaignCard key={c.id} campaign={c} />
          ))}
        </div>
      )}
    </AppLayout>
  );
}
