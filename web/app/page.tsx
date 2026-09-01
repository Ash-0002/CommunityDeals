// Landing page — root route "/"
// Redirects logged-in users to /dashboard, shows marketing page otherwise.

import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function LandingPage() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-emerald-50">
      {/* Nav */}
      <nav className="mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
        <span className="text-xl font-bold text-primary-600">🏘 CommunityDeals</span>
        <div className="flex gap-3">
          <Link href="/auth/login">
            <Button variant="outline" size="sm">Log in</Button>
          </Link>
        </div>
      </nav>

      {/* Hero */}
      <section className="mx-auto max-w-3xl px-6 py-20 text-center">
        <span className="mb-4 inline-block rounded-full bg-primary-100 px-4 py-1 text-sm font-medium text-primary-700">
          For housing societies &amp; apartments 🏢
        </span>
        <h1 className="mb-6 text-4xl font-bold leading-tight text-gray-900 sm:text-5xl">
          The more neighbours who join,
          <br />
          <span className="text-primary-500">the less everyone pays.</span>
        </h1>
        <p className="mx-auto mb-10 max-w-xl text-lg text-gray-600">
          Organise group deals for car washing, AC servicing, pest control and more.
          Share one WhatsApp link — watch the price drop as your neighbours join.
        </p>
        <div className="flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <Link href="/auth/login">
            <Button size="lg">Get started free →</Button>
          </Link>
        </div>
      </section>

      {/* How it works */}
      <section className="mx-auto max-w-4xl px-6 pb-24">
        <h2 className="mb-10 text-center text-2xl font-bold text-gray-800">How it works</h2>
        <div className="grid gap-6 sm:grid-cols-3">
          {[
            {
              step: "1",
              title: "Admin creates a campaign",
              desc: "Set the service, date, and group pricing tiers (e.g. 10+ people → ₹500).",
              icon: "📋",
            },
            {
              step: "2",
              title: "Share on WhatsApp",
              desc: "One link. Anyone who clicks it sees the campaign and the live count.",
              icon: "💬",
            },
            {
              step: "3",
              title: "Price drops as people join",
              desc: "Hit the next tier → everyone gets the better price automatically.",
              icon: "🎉",
            },
          ].map((item) => (
            <div key={item.step} className="card p-6 text-center">
              <div className="mb-4 text-4xl">{item.icon}</div>
              <h3 className="mb-2 font-semibold text-gray-900">{item.title}</h3>
              <p className="text-sm text-gray-600">{item.desc}</p>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}
