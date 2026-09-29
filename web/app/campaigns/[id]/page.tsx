"use client";

// Authenticated campaign detail — /campaigns/[id]
// Editorial layout; the join / cancel action lives in the sticky join card.

import { useCallback, useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ChevronLeft } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { CampaignEditorial } from "@/components/campaign/campaign-detail-body";
import { Button } from "@/components/ui/button";
import { api, APIError } from "@/lib/api";
import type { CampaignResponse, CommunityResponse, ParticipantResponse } from "@/lib/types";

export default function CampaignDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [campaign, setCampaign] = useState<CampaignResponse | null>(null);
  const [participants, setParticipants] = useState<ParticipantResponse[]>([]);
  const [community, setCommunity] = useState<CommunityResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [acting, setActing] = useState(false);
  const [toast, setToast] = useState("");
  const [confirmingCancel, setConfirmingCancel] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    setError("");
    Promise.all([api.campaigns.get(id), api.campaigns.participants(id)])
      .then(([c, p]) => {
        setCampaign(c);
        setParticipants(p.participants);
        // The community carries the society name + location. Best effort —
        // a viewer who isn't a member simply gets no location block.
        api.communities
          .get(c.community_id)
          .then(setCommunity)
          .catch(() => setCommunity(null));
      })
      .catch((err) => setError(err instanceof APIError ? err.message : "Failed to load campaign"))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  async function doToggleJoin() {
    if (!campaign) return;
    setActing(true);
    setToast("");
    setConfirmingCancel(false);
    try {
      if (campaign.is_joined) {
        await api.campaigns.leave(campaign.id);
        setToast("You left this deal");
      } else {
        const r = await api.campaigns.join(campaign.id);
        setToast(r.message);
      }
      load();
    } catch (err) {
      setToast(err instanceof APIError ? err.message : "Something went wrong");
    } finally {
      setActing(false);
    }
  }

  function handleActionClick() {
    if (campaign?.is_joined) {
      setConfirmingCancel(true);
    } else {
      doToggleJoin();
    }
  }

  const closed = campaign?.status === "CANCELLED" || campaign?.status === "COMPLETED";

  return (
    <AppLayout>
      <button
        onClick={() => router.back()}
        className="mb-4 inline-flex items-center gap-1 rounded-full border border-border bg-card px-3.5 py-1.5 text-[13px] font-semibold text-muted transition-colors hover:text-ink"
      >
        <ChevronLeft className="h-4 w-4" />
        Back
      </button>

      {loading ? (
        <p className="text-sm text-muted">Loading…</p>
      ) : error || !campaign ? (
        <p className="text-sm text-red-500">{error || "Campaign not found"}</p>
      ) : (
        <CampaignEditorial
          campaign={campaign}
          participants={participants}
          community={community}
          action={
            confirmingCancel ? (
              // Confirmation step before giving up a locked-in spot.
              <div className="rounded-2xl border border-red-200 p-3.5 dark:border-red-900/50">
                <p className="mb-3 text-[13px] leading-relaxed text-ink">
                  You&apos;ll leave this deal and lose your locked-in price. You can rejoin later if
                  it&apos;s still open.
                </p>
                <div className="flex flex-col gap-2">
                  <Button fullWidth variant="outline" onClick={() => setConfirmingCancel(false)}>
                    Keep my spot
                  </Button>
                  <Button fullWidth variant="danger" loading={acting} onClick={doToggleJoin}>
                    Cancel spot
                  </Button>
                </div>
              </div>
            ) : (
              <Button
                fullWidth
                size="lg"
                variant={campaign.is_joined ? "danger" : "primary"}
                loading={acting}
                disabled={closed}
                onClick={handleActionClick}
              >
                {closed ? "Deal closed" : campaign.is_joined ? "Cancel my spot" : "Join this deal"}
              </Button>
            )
          }
          actionNote={
            toast ? (
              <span className="font-semibold text-ink">{toast}</span>
            ) : (
              "No payment now · Pay only when the group is confirmed"
            )
          }
        />
      )}
    </AppLayout>
  );
}
