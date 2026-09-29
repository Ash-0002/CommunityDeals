// Public campaign page — /c/[slug]
// No auth required. Shareable via WhatsApp. Mirrors GET /c/:slug on the backend.

import { cookies } from "next/headers";
import { notFound } from "next/navigation";
import { CampaignPublicView } from "@/components/campaign/campaign-public-view";
import type { APIResponse, CampaignResponse, CommunityResponse } from "@/lib/types";

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

/**
 * The community (society name + city/state) is only readable with a session.
 * A signed-in visitor opening the share link gets the location block; an
 * anonymous one simply doesn't — no new API surface, no hard failure.
 */
async function getCommunity(id: string): Promise<CommunityResponse | null> {
  const token = cookies().get("cp_access_token")?.value;
  if (!token) return null;
  try {
    const res = await fetch(`${API_URL}/api/v1/communities/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const body = (await res.json()) as APIResponse<CommunityResponse>;
    return body.success && body.data ? body.data : null;
  } catch {
    return null;
  }
}

export default async function CampaignPublicPage({ params }: { params: { slug: string } }) {
  const campaign = await getCampaign(params.slug);
  if (!campaign) notFound();
  const community = await getCommunity(campaign.community_id);
  return <CampaignPublicView campaign={campaign} community={community} />;
}

export async function generateMetadata({ params }: { params: { slug: string } }) {
  const campaign = await getCampaign(params.slug);
  if (!campaign) return { title: "Deal not found" };
  return {
    title: campaign.title,
    description:
      campaign.description ||
      `Join your neighbours for ${campaign.service_name} — the price drops as more people join.`,
    openGraph: {
      title: campaign.title,
      description: campaign.description,
      images: campaign.image_url ? [campaign.image_url] : undefined,
    },
  };
}
