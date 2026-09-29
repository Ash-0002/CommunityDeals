"use client";

// Authenticated campaign detail — /campaigns/[id]
// Full campaign view with a working join/cancel action.

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ChevronLeft } from "lucide-react";
import { AppLayout } from "@/components/layout/app-layout";
import { CampaignDetailBody } from "@/components/campaign/campaign-detail-body";
import { Button } from "@/components/ui/button";
import { api, APIError } from "@/lib/api";
import type { CampaignResponse, ParticipantResponse } from "@/lib/types";

export default function CampaignDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [campaign, setCampaign] = useState<CampaignResponse | null>(null);
  const [participants, setParticipants] = useState<ParticipantResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [acting, setActing] = useState(false);
  const [toast, setToast] = useState("");
  const [confirmingCancel, setConfirmingCancel] = useState(false);

  function load() {
    setLoading(true);
    setError("");
    Promise.all([api.campaigns.get(id), api.campaigns.participants(id)])
      .then(([c, p]) => {
        setCampaign(c);
        setParticipants(p.participants);
      })
      .catch((err) => setError(err instanceof APIError ? err.message : "Failed to load campaign"))
      .finally(() => setLoading(false));
  }

  useEffect(load, [id]);

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
      <div className="flex items-center gap-3 border-b border-border bg-card px-6 py-4">
        <button onClick={() => router.back()} className="rounded-md p-1 text-muted hover:bg-surface hover:text-ink">
          <ChevronLeft className="h-4 w-4" />
        </button>
        <h1 className="text-sm font-semibold text-ink">Deal details</h1>
      </div>

      <div className="flex-1 overflow-y-auto px-6 py-6">
        <div className="mx-auto max-w-xl">
          {loading ? (
            <p className="text-sm text-muted">Loading…</p>
          ) : error || !campaign ? (
            <p className="text-sm text-red-500">{error || "Campaign not found"}</p>
          ) : (
            <>
              <CampaignDetailBody campaign={campaign} participants={participants} />

              {toast && <p className="mt-4 text-center text-sm font-medium text-ink">{toast}</p>}

              {confirmingCancel ? (
                <div className="mt-6 rounded-xl border border-border bg-card p-4">
                  <p className="mb-3 text-sm text-ink">
                    You&apos;ll leave &quot;{campaign.title}&quot; and lose your locked-in price. You can rejoin
                    later if the deal is still open.
                  </p>
                  <div className="flex gap-2">
                    <Button fullWidth variant="outline" onClick={() => setConfirmingCancel(false)}>
                      Keep my spot
                    </Button>
                    <Button fullWidth variant="danger" loading={acting} onClick={doToggleJoin}>
                      Cancel spot
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="mt-6">
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
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </AppLayout>
  );
}
