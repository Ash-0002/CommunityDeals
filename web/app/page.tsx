// Landing page — root route "/"

import Link from "next/link";
import { ClipboardList, Share2, TrendingDown } from "lucide-react";
import { Button } from "@/components/ui/button";

const STEPS = [
  {
    icon: ClipboardList,
    title: "Admin creates a campaign",
    desc: "Set the service, date, and group pricing tiers — e.g. 10+ people unlocks ₹500.",
  },
  {
    icon: Share2,
    title: "Share one link",
    desc: "Anyone who opens it sees the campaign and the live participant count.",
  },
  {
    icon: TrendingDown,
    title: "Price drops as people join",
    desc: "Hit the next tier and everyone gets the better price automatically.",
  },
];

export default function LandingPage() {
  return (
    <main className="min-h-screen bg-surface">
      <nav className="mx-auto flex max-w-4xl items-center justify-between px-6 py-5">
        <span className="flex items-center gap-2 text-sm font-bold text-ink">
          <span className="flex h-7 w-7 items-center justify-center rounded-md bg-primary-600 text-xs font-bold text-white">C</span>
          CommunityDeals
        </span>
        <Link href="/auth/login">
          <Button variant="outline" size="sm">Log in</Button>
        </Link>
      </nav>

      <section className="mx-auto max-w-2xl px-6 py-20 text-center">
        <span className="mb-5 inline-block rounded-full border border-border bg-card px-3.5 py-1 text-xs font-medium text-muted">
          For housing societies &amp; apartments
        </span>
        <h1 className="mb-5 text-4xl font-bold leading-tight text-ink sm:text-5xl">
          The more neighbours who join,
          <br />
          <span className="text-primary-600">the less everyone pays.</span>
        </h1>
        <p className="mx-auto mb-9 max-w-lg text-base leading-relaxed text-muted">
          Organise group deals for AC servicing, pest control, and more. Share one link —
          watch the price drop as your neighbours join.
        </p>
        <Link href="/auth/login">
          <Button size="lg">Get started</Button>
        </Link>
      </section>

      <section className="mx-auto max-w-3xl px-6 pb-24">
        <h2 className="mb-8 text-center text-lg font-bold text-ink">How it works</h2>
        <div className="grid gap-4 sm:grid-cols-3">
          {STEPS.map((step) => (
            <div key={step.title} className="rounded-xl border border-border bg-card p-5">
              <step.icon className="mb-3 h-5 w-5 text-primary-600" strokeWidth={2} />
              <h3 className="mb-1.5 text-sm font-semibold text-ink">{step.title}</h3>
              <p className="text-[13px] leading-relaxed text-muted">{step.desc}</p>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}
