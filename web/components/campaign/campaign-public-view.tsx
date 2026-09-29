// Public share page view — what someone sees after tapping a WhatsApp link.
// Phone-first: designed to read well at 375px, then widen into the same
// editorial two-column layout as the authenticated detail page.

import Link from "next/link";
import type { CampaignResponse, CommunityResponse } from "@/lib/types";
import { formatPaise } from "@/lib/format";
import { CampaignEditorial } from "./campaign-detail-body";
import { Button } from "@/components/ui/button";

export function CampaignPublicView({
  campaign,
  community = null,
}: {
  campaign: CampaignResponse;
  community?: CommunityResponse | null;
}) {
  const joinHref = `/auth/login?redirect=/campaigns/${campaign.id}`;

  return (
    <div className="min-h-screen bg-surface">
      {/* Same floating glass bar as the app, minus the authed nav */}
      <header className="sticky top-0 z-40 px-3 pt-3 sm:px-4 sm:pt-5">
        <div className="glass mx-auto flex max-w-5xl items-center gap-3 rounded-full border border-border px-2.5 py-2 shadow-card-hover sm:px-3">
          <Link href="/" className="flex min-w-0 items-center gap-2">
            <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-primary-600 text-sm font-black text-white">
              C
            </span>
            <span className="truncate text-[15px] font-extrabold tracking-tight text-ink">
              CommunityDeals
            </span>
          </Link>
          <Link
            href="/auth/login"
            className="ml-auto shrink-0 rounded-full bg-primary-600 px-4 py-2 text-[13px] font-bold text-white transition-colors hover:bg-primary-700"
          >
            Log in
          </Link>
        </div>
      </header>

      <main className="mx-auto w-full max-w-5xl px-4 pb-16 pt-6 sm:px-6 sm:pt-8">
        <CampaignEditorial
          campaign={campaign}
          community={community}
          action={
            <Link href={joinHref} className="block">
              <Button fullWidth size="lg">
                Join — {formatPaise(campaign.current_price)}
              </Button>
            </Link>
          }
          actionNote="No payment now · Pay only when the group is confirmed"
        />

        {/* Warm closing note for first-time visitors */}
        <section className="mt-8 rounded-3xl border border-border bg-card p-6 text-center sm:p-8">
          <h2 className="text-lg font-extrabold tracking-tight text-ink sm:text-xl">
            How group deals work
          </h2>
          <p className="mx-auto mt-2 max-w-md text-sm leading-relaxed text-muted">
            Neighbours book the same service together. The more flats that join, the lower the price
            drops — for everyone, automatically.
          </p>
          <Link href={joinHref} className="mt-5 inline-block">
            <Button size="lg">Join your neighbours</Button>
          </Link>
        </section>
      </main>
    </div>
  );
}
