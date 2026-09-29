"use client";

import Link from "next/link";
import { Users } from "lucide-react";
import type { CampaignResponse } from "@/lib/types";
import { formatPaise } from "@/lib/format";
import { CampaignDetailBody } from "./campaign-detail-body";
import { Button } from "@/components/ui/button";

export function CampaignPublicView({ campaign }: { campaign: CampaignResponse }) {
  return (
    <div className="min-h-screen bg-surface">
      <header className="border-b border-border bg-card px-6 py-3.5">
        <div className="mx-auto flex max-w-3xl items-center justify-between">
          <Link href="/" className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary-600 text-xs font-bold text-white">
              C
            </div>
            <span className="text-sm font-bold text-ink">CommunityDeals</span>
          </Link>
          <Link href="/auth/login" className="rounded-lg bg-primary-600 px-4 py-1.5 text-xs font-semibold text-white hover:bg-primary-700">
            Log in
          </Link>
        </div>
      </header>

      <div className="mx-auto max-w-3xl px-6 py-8">
        <div className="grid grid-cols-[1fr_300px] items-start gap-6">
          <div>
            <CampaignDetailBody campaign={campaign} />
          </div>

          {/* Sticky join panel */}
          <div className="sticky top-6 rounded-xl border border-border bg-card p-5">
            <div className="mb-3 flex items-center justify-center gap-2 text-center">
              <Users className="h-4 w-4 text-muted" />
              <span className="text-sm font-medium text-muted">
                {campaign.participant_count} of {campaign.min_participants} joined
              </span>
            </div>

            <Link href={`/auth/login?redirect=/campaigns/${campaign.id}`}>
              <Button fullWidth size="lg">
                Join — {formatPaise(campaign.current_price)}
              </Button>
            </Link>

            <p className="mt-3 text-center text-[11px] text-muted">
              No payment now · Pay only when the group is confirmed
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
