"use client";

// Floating bottom pill nav — phones only (hidden at md and up).
// Home · Deals · + (start a deal) · Profile

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Plus, type LucideIcon } from "lucide-react";
import { NAV_ITEMS, isActive } from "./nav-items";

export function BottomNav() {
  const pathname = usePathname();

  return (
    <div
      className="fixed inset-x-0 bottom-0 z-40 flex justify-center px-4 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-2 md:hidden"
      style={{ pointerEvents: "none" }}
    >
      <nav
        className="glass flex items-center gap-1 rounded-full border border-border p-1.5 shadow-card-hover"
        style={{ pointerEvents: "auto" }}
      >
        {NAV_ITEMS.slice(0, 2).map((item) => (
          <NavPill key={item.href} {...item} active={isActive(pathname, item.href)} />
        ))}

        <Link
          href="/campaigns/new"
          aria-label="Start a deal"
          className="mx-0.5 flex h-11 w-11 items-center justify-center rounded-full bg-primary-600 text-white shadow-md transition-colors hover:bg-primary-700"
        >
          <Plus className="h-5 w-5" strokeWidth={3} />
        </Link>

        {NAV_ITEMS.slice(2).map((item) => (
          <NavPill key={item.href} {...item} active={isActive(pathname, item.href)} />
        ))}
      </nav>
    </div>
  );
}

function NavPill({
  href,
  label,
  icon: Icon,
  active,
}: {
  href: string;
  label: string;
  icon: LucideIcon;
  active: boolean;
}) {
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className={`flex min-w-[68px] flex-col items-center gap-0.5 rounded-full px-3 py-2 transition-colors ${
        active ? "bg-primary-600 text-white" : "text-muted"
      }`}
    >
      <Icon className="h-[18px] w-[18px]" strokeWidth={2.2} />
      <span className="text-[10.5px] font-bold leading-none">{label}</span>
    </Link>
  );
}
