// User dashboard — /dashboard
// Shows the user's joined campaigns and upcoming services.

import { AppLayout } from "@/components/layout/app-layout";
import Link from "next/link";

const MY_CAMPAIGNS = [
  {
    id: "1",
    title: "AC Service — This Sunday",
    community: "Green Valley Society",
    joinedDaysAgo: 2,
    price: 99900,
    status: "PUBLISHED" as const,
    slug: "ac-service-green-valley-aug31",
    icon: "❄️",
    statusLabel: "2 more needed",
    statusColor: "text-amber-500",
  },
  {
    id: "3",
    title: "Pest Control — Full Society",
    community: "Green Valley Society",
    joinedDaysAgo: 7,
    price: 80000,
    status: "CONFIRMED" as const,
    slug: "pest-control-green-valley-sep12",
    icon: "🐛",
    statusLabel: "Group locked ✓",
    statusColor: "text-primary-500",
  },
  {
    id: "5",
    title: "Water Softener Install",
    community: "Sunrise Heights",
    joinedDaysAgo: 20,
    price: 240000,
    status: "COMPLETED" as const,
    slug: "water-softener-sunrise-aug15",
    icon: "🚿",
    statusLabel: "Completed",
    statusColor: "text-muted",
  },
];

const UPCOMING = [
  { day: "31", month: "Aug", title: "AC Service", meta: "10:00 AM – 1:00 PM · Green Valley" },
  { day: "12", month: "Sep", title: "Pest Control", meta: "9:00 AM – 5:00 PM · Green Valley" },
];

const COMMUNITIES = [
  { name: "Green Valley Society", location: "Sector 14, Gurgaon", members: 248, active: 3, done: 14 },
  { name: "Sunrise Heights", location: "Noida", members: 92, active: 1, done: 6 },
];

function formatPrice(paise: number) {
  return `₹${(paise / 100).toLocaleString("en-IN")}`;
}

export default function DashboardPage() {
  return (
    <AppLayout>
      {/* Top bar */}
      <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
        <div>
          <h1 className="text-base font-extrabold text-ink">My Dashboard</h1>
          <p className="text-xs font-medium text-muted">Rahul Mehta · Green Valley Society</p>
        </div>
        <Link
          href="/campaigns"
          className="rounded-lg bg-primary-600 px-3.5 py-1.5 text-xs font-bold text-white hover:bg-primary-700"
        >
          + Join a Campaign
        </Link>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-6 py-5">
        <div className="grid grid-cols-[1fr_272px] gap-5 items-start">

          {/* Left */}
          <div>
            <h2 className="mb-3 text-sm font-extrabold text-ink">My Campaigns</h2>
            <div className="flex flex-col gap-2.5">
              {MY_CAMPAIGNS.map((c) => (
                <Link
                  key={c.id}
                  href={`/c/${c.slug}`}
                  className={`flex items-center gap-3.5 rounded-xl bg-card border border-border p-3.5 shadow-card hover:shadow-card-hover transition-shadow ${
                    c.status === "COMPLETED" ? "opacity-60" : ""
                  }`}
                >
                  <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-surface text-xl">
                    {c.icon}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[13px] font-bold text-ink">{c.title}</p>
                    <p className="text-[11px] font-medium text-muted">
                      {c.community} · Joined {c.joinedDaysAgo}d ago
                    </p>
                  </div>
                  <div className="flex-shrink-0 text-right">
                    <p className="num text-sm font-extrabold text-primary-500">
                      {formatPrice(c.price)}
                    </p>
                    <p className={`text-[10px] font-bold ${c.statusColor}`}>{c.statusLabel}</p>
                  </div>
                </Link>
              ))}
            </div>

            {/* Upcoming services */}
            <h2 className="mb-3 mt-6 text-sm font-extrabold text-ink">Upcoming Services</h2>
            <div className="overflow-hidden rounded-xl bg-card border border-border shadow-card">
              {UPCOMING.map((item, i) => (
                <div
                  key={i}
                  className={`flex items-center gap-3.5 px-4 py-3 ${
                    i !== UPCOMING.length - 1 ? "border-b border-border" : ""
                  }`}
                >
                  <div className="w-9 flex-shrink-0 text-center">
                    <p className="num text-base font-extrabold leading-none text-ink">{item.day}</p>
                    <p className="text-[9px] font-bold uppercase tracking-wide text-muted">{item.month}</p>
                  </div>
                  <div className="w-px self-stretch bg-border" />
                  <div>
                    <p className="text-[13px] font-bold text-ink">{item.title}</p>
                    <p className="text-[11px] text-muted">{item.meta}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Right: communities */}
          <div>
            <h2 className="mb-3 text-sm font-extrabold text-ink">My Communities</h2>
            <div className="flex flex-col gap-3">
              {COMMUNITIES.map((comm) => (
                <div key={comm.name} className="rounded-xl bg-card border border-border p-4 shadow-card">
                  <p className="text-[13px] font-bold text-ink">{comm.name}</p>
                  <p className="mt-0.5 flex items-center gap-1 text-[11px] font-medium text-muted">
                    👥 {comm.members} members · {comm.location}
                  </p>
                  <div className="mt-3 grid grid-cols-2 gap-2">
                    <div className="rounded-lg bg-surface p-2 text-center">
                      <p className="num text-base font-extrabold text-ink">{comm.active}</p>
                      <p className="text-[9px] font-semibold uppercase tracking-wide text-muted">Active</p>
                    </div>
                    <div className="rounded-lg bg-surface p-2 text-center">
                      <p className="num text-base font-extrabold text-ink">{comm.done}</p>
                      <p className="text-[9px] font-semibold uppercase tracking-wide text-muted">Done</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

        </div>
      </div>
    </AppLayout>
  );
}
