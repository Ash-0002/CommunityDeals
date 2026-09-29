// Public campaign page — /c/[slug]
// No auth required. Shareable via WhatsApp. Mirrors GET /c/:slug on the backend.

import { notFound } from "next/navigation";
import { CampaignPublicView } from "@/components/campaign/campaign-public-view";
import type { APIResponse, CampaignResponse } from "@/lib/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function getCampaign(slug: string): Promise<CampaignResponse | null> {
  try {
    const res = await fetch(`${API_URL}/c/${slug}`, { next: { revalidate: 30 } });
    if (!res.ok) return null;
    const body = (await res.json()) as APIResponse<CampaignResponse>;
    return body.success && body.data ? body.data : null;
  } catch {
    return null;
  }
}

export default async function CampaignPublicPage({ params }: { params: { slug: string } }) {
  const campaign = await getCampaign(params.slug);
  if (!campaign) notFound();
  return <CampaignPublicView campaign={campaign} />;
}
