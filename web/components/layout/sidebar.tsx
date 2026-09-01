"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ThemeToggle } from "@/components/theme-toggle";

const navItems = [
  {
    section: "Discover",
    links: [
      { href: "/campaigns",   label: "Campaigns",   icon: "⊞" },
      { href: "/communities", label: "Communities", icon: "⬡" },
    ],
  },
  {
    section: "My Activity",
    links: [
      { href: "/dashboard", label: "Dashboard",  icon: "⌂" },
      { href: "/bookings",  label: "My Bookings", icon: "📅" },
    ],
  },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex w-56 flex-shrink-0 flex-col bg-sidebar px-3 py-5">
      {/* Logo */}
      <div className="mb-5 flex items-center gap-2.5 px-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-600 text-sm font-black text-white">
          C
        </div>
        <span className="text-[15px] font-extrabold text-white">CommunityDeals</span>
      </div>

      {/* Nav links */}
      <nav className="flex flex-col gap-0.5">
        {navItems.map((group) => (
          <div key={group.section}>
            <p className="mb-1 mt-4 px-2 text-[9px] font-bold uppercase tracking-widest text-white/30">
              {group.section}
            </p>
            {group.links.map((link) => {
              const active = pathname.startsWith(link.href);
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] font-semibold transition-colors ${
                    active
                      ? "bg-primary-600/20 text-green-400"
                      : "text-white/50 hover:bg-white/5 hover:text-white/80"
                  }`}
                >
                  <span className="text-base">{link.icon}</span>
                  {link.label}
                </Link>
              );
            })}
          </div>
        ))}
      </nav>

      {/* Bottom */}
      <div className="mt-auto border-t border-white/10 pt-3">
        {/* Theme toggle */}
        <div className="mb-2 flex items-center justify-between px-2">
          <span className="text-[10px] font-semibold text-white/30">Appearance</span>
          <ThemeToggle />
        </div>

        {/* User */}
        <Link
          href="/profile"
          className="flex items-center gap-2 rounded-lg px-2 py-2 text-white/50 hover:bg-white/5 hover:text-white/80"
        >
          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-primary-500 to-sky-500 text-xs font-bold text-white">
            R
          </div>
          <div>
            <p className="text-[12px] font-bold text-white/70">Rahul Mehta</p>
            <p className="text-[10px] text-white/30">Green Valley</p>
          </div>
        </Link>
      </div>
    </aside>
  );
}
