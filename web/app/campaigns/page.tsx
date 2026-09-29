"use client";

// Campaigns discovery page — /campaigns
// Lists real campaigns from the API with a status filter, mirroring the
// mobile app's Deals tab.

import { useEffect, useState } from "react";
import Link from "next/link";
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
      <div className="border-b border-border bg-card px-6 py-4">
        <div className="mb-3 flex items-center justify-between">
          <h1 className="text-base font-bold text-ink">Deals</h1>
          <Link
            href="/campaigns/new"
            className="rounded-lg bg-primary-600 px-3.5 py-1.5 text-xs font-semibold text-white hover:bg-primary-700"
          >
            + Start a deal
          </Link>
        </div>
        <div className="flex gap-2">
          {FILTERS.map((f) => (
            <button
              key={f.label}
              onClick={() => setActive(f)}
              className={`rounded-full px-3 py-1 text-xs font-semibold transition-colors ${
                active.label === f.label
                  ? "bg-primary-600 text-white"
                  : "border border-border text-muted hover:text-ink"
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-6 py-6">
        {loading ? (
          <p className="text-sm text-muted">Loading…</p>
        ) : error ? (
          <p className="text-sm text-red-500">{error}</p>
        ) : campaigns.length === 0 ? (
          <p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted">
            Nothing here yet.
          </p>
        ) : (
          <div className="grid grid-cols-3 gap-4">
            {campaigns.map((c) => (
              <CampaignCard key={c.id} campaign={c} />
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  );
}
