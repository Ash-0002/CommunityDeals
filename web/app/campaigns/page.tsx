// Campaigns discovery page — /campaigns
// Shows all active campaigns in the user's community.

import { AppLayout } from "@/components/layout/app-layout";
import { CampaignCard } from "@/components/campaign/campaign-card";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import type { CampaignResponse } from "@/lib/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function StatCard({ label, value, sub }: { label: string; value: string; sub: string }) {
  return (
    <div className="rounded-xl bg-card border border-border p-4 shadow-card">
      <p className="mb-1.5 text-[10px] font-bold uppercase tracking-wider text-muted">{label}</p>
      <p className="num text-2xl font-extrabold text-ink">{value}</p>
      <p className="mt-1 text-[11px] font-semibold text-muted">{sub}</p>
    </div>
  );
}

const PLACEHOLDER_CAMPAIGNS: CampaignResponse[] = [
  {
    id: "1",
    community_id: "c1",
    community_name: "Green Valley Society",
    created_by_id: "u1",
    title: "AC Service — This Sunday",
    description: "Annual AC servicing for all flats.",
    service_type: "AC_SERVICE",
    slug: "ac-service-green-valley-aug31",
    status: "PUBLISHED",
    min_participants: 20,
    max_participants: null,
    participant_count: 18,
    current_price: 99900,
    pricing_tiers: [],
    next_tier: null,
    participants_to_next_tier: 2,
    is_joined: false,
    service_date: "2026-08-31T10:00:00Z",
    campaign_end_date: null,
    location_name: "Green Valley Society",
    location_address: "Sector 14, Gurgaon",
    city: "Gurgaon",
    pin_code: "122001",
    image_url: "",
    vendor_name: "CoolTech Services",
    share_url: "http://localhost:3000/c/ac-service-green-valley-aug31",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "2",
    community_id: "c1",
    community_name: "Green Valley Society",
    created_by_id: "u1",
    title: "Water Tank Deep Cleaning",
    description: "Full society water tank cleaning.",
    service_type: "WATER_TANK",
    slug: "water-tank-green-valley-sep6",
    status: "PUBLISHED",
    min_participants: 15,
    max_participants: null,
    participant_count: 7,
    current_price: 120000,
    pricing_tiers: [],
    next_tier: null,
    participants_to_next_tier: 8,
    is_joined: false,
    service_date: "2026-09-06T09:00:00Z",
    campaign_end_date: null,
    location_name: "Green Valley Society",
    location_address: "Sector 14, Gurgaon",
    city: "Gurgaon",
    pin_code: "122001",
    image_url: "",
    vendor_name: "AquaClean",
    share_url: "",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "3",
    community_id: "c1",
    community_name: "Green Valley Society",
    created_by_id: "u1",
    title: "Pest Control — Full Society",
    description: "Comprehensive pest control.",
    service_type: "PEST_CONTROL",
    slug: "pest-control-green-valley-sep12",
    status: "CONFIRMED",
    min_participants: 25,
    max_participants: null,
    participant_count: 25,
    current_price: 80000,
    pricing_tiers: [],
    next_tier: null,
    participants_to_next_tier: 0,
    is_joined: true,
    service_date: "2026-09-12T09:00:00Z",
    campaign_end_date: null,
    location_name: "Green Valley Society",
    location_address: "Sector 14, Gurgaon",
    city: "Gurgaon",
    pin_code: "122001",
    image_url: "",
    vendor_name: "PestAway Pro",
    share_url: "",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export default async function CampaignsPage() {
  const campaigns = PLACEHOLDER_CAMPAIGNS;

  return (
    <AppLayout>
      {/* Top bar */}
      <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
        <div>
          <h1 className="text-base font-extrabold text-ink">Active Campaigns</h1>
          <p className="text-xs font-medium text-muted">
            Green Valley Society · {campaigns.length} open deals
          </p>
        </div>
        <Link href="/campaigns/new">
          <Button size="sm">+ New Campaign</Button>
        </Link>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto px-6 py-5">
        {/* Stats */}
        <div className="mb-6 grid grid-cols-4 gap-3">
          <StatCard label="Members"     value="248"                     sub="12 joined this week" />
          <StatCard label="Active"      value={String(campaigns.length)} sub="campaigns open"      />
          <StatCard label="Total Saved" value="₹1.2L"                   sub="across community"    />
          <StatCard label="Completed"   value="14"                      sub="campaigns this year" />
        </div>

        {/* Campaign grid */}
        <h2 className="mb-3.5 text-sm font-extrabold text-ink">Open Deals</h2>
        <div className="grid grid-cols-3 gap-4">
          {campaigns.map((c) => (
            <CampaignCard key={c.id} campaign={c} />
          ))}
        </div>
      </div>
    </AppLayout>
  );
}
