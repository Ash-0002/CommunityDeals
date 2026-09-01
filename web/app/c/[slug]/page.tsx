// Public campaign page — /c/[slug]
// No auth required. Shareable via WhatsApp.

import { CampaignPublicView } from "@/components/campaign/campaign-public-view";
import type { CampaignResponse } from "@/lib/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// Placeholder shown when API isn't reachable
const PLACEHOLDER: CampaignResponse = {
  id: "1",
  community_id: "c1",
  community_name: "Green Valley Society",
  created_by_id: "u1",
  title: "AC Service — This Sunday",
  description:
    "Annual AC servicing for all flats. CoolTech Services will visit each flat in sequence. One AC per flat included. Please ensure someone is home between 10 AM and 1 PM.",
  service_type: "AC_SERVICE",
  slug: "ac-service-green-valley-aug31",
  status: "PUBLISHED",
  min_participants: 20,
  max_participants: null,
  participant_count: 18,
  current_price: 99900,
  pricing_tiers: [
    { min_count: 1,  price: 99900, label: "Individual" },
    { min_count: 10, price: 69900, label: "Group"       },
    { min_count: 25, price: 39900, label: "Society"     },
  ],
  next_tier: { min_count: 25, price: 39900, label: "Society" },
  participants_to_next_tier: 7,
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
};

async function getCampaign(slug: string): Promise<CampaignResponse> {
  try {
    const res = await fetch(`${API_URL}/api/v1/campaigns/${slug}`, {
      next: { revalidate: 30 },
    });
    if (!res.ok) return PLACEHOLDER;
    const body = await res.json();
    return body.data ?? PLACEHOLDER;
  } catch {
    return PLACEHOLDER;
  }
}

export default async function CampaignPublicPage({
  params,
}: {
  params: { slug: string };
}) {
  const campaign = await getCampaign(params.slug);
  return <CampaignPublicView campaign={campaign} />;
}
