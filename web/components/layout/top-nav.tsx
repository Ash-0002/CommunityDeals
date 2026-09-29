"use client";

// Floating glass top bar — replaces the old fixed sidebar.
//
//   ≥ md : logo + wordmark · centred pill nav · "+ Start a deal" · theme · avatar
//   < md : logo + wordmark · theme · avatar   (nav moves to <BottomNav />)
//
// The avatar opens a small menu that holds "View profile" and "Log out", so
// log out stays reachable from every authenticated page.

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { LogOut, Plus, UserRound } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { NAV_ITEMS, isActive } from "./nav-items";
import { useUser, userInitials } from "@/lib/use-user";
import { getRefreshToken, clearTokens } from "@/lib/auth";
import { api } from "@/lib/api";

export function TopNav() {
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useUser();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close the avatar menu on outside click / Escape
  useEffect(() => {
    if (!menuOpen) return;
    function onDown(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuOpen(false);
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setMenuOpen(false);
    }
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [menuOpen]);

  // Close the menu whenever the route changes
  useEffect(() => setMenuOpen(false), [pathname]);

  async function handleLogout() {
    const refreshToken = getRefreshToken();
    try {
      if (refreshToken) await api.auth.logout(refreshToken);
    } catch {
      // best-effort — clear the local session regardless
    } finally {
      clearTokens();
      router.replace("/auth/login");
    }
  }

  return (
    <header className="sticky top-0 z-40 px-3 pt-3 sm:px-4 sm:pt-5">
      <div className="glass mx-auto flex max-w-6xl items-center gap-2 rounded-full border border-border px-2.5 py-2 shadow-card-hover sm:gap-3 sm:px-3">
        {/* Logo + wordmark */}
        <Link href="/dashboard" className="flex shrink-0 items-center gap-2 rounded-full px-1 py-0.5">
          <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-600 text-sm font-black text-white shadow-sm">
            C
          </span>
          <span className="hidden text-[15px] font-extrabold tracking-tight text-ink sm:inline">
            CommunityDeals
          </span>
        </Link>

        {/* Centred pill nav — md and up */}
        <nav className="mx-auto hidden items-center gap-1 rounded-full bg-surface p-1 md:flex">
          {NAV_ITEMS.map((item) => {
            const active = isActive(pathname, item.href);
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={active ? "page" : undefined}
                className={`rounded-full px-4 py-1.5 text-[13px] font-semibold transition-colors ${
                  active
                    ? "bg-primary-600 text-white shadow-sm"
                    : "text-muted hover:text-ink"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* Right cluster */}
        <div className="ml-auto flex items-center gap-1.5 md:ml-0 md:gap-2">
          <Link
            href="/campaigns/new"
            className="hidden items-center gap-1.5 rounded-full bg-primary-600 px-4 py-2 text-[13px] font-bold text-white transition-colors hover:bg-primary-700 md:inline-flex"
          >
            <Plus className="h-3.5 w-3.5" strokeWidth={3} />
            Start a deal
          </Link>

          <ThemeToggle />

          <div className="relative" ref={menuRef}>
            <button
              onClick={() => setMenuOpen((v) => !v)}
              aria-haspopup="menu"
              aria-expanded={menuOpen}
              aria-label="Account menu"
              className={`flex h-9 w-9 items-center justify-center rounded-full bg-primary-600 text-[12px] font-black text-white ring-offset-2 ring-offset-card transition-shadow hover:ring-2 hover:ring-primary-500/50 ${
                isActive(pathname, "/profile") ? "ring-2 ring-primary-500/60" : ""
              }`}
            >
              {userInitials(user?.name)}
            </button>

            {menuOpen && (
              <div
                role="menu"
                className="absolute right-0 top-11 w-56 overflow-hidden rounded-2xl border border-border bg-card p-1.5 shadow-card-hover"
              >
                <div className="px-3 py-2">
                  <p className="truncate text-[13px] font-bold text-ink">
                    {user?.name || "My profile"}
                  </p>
                  <p className="truncate text-[11.5px] text-muted">{user?.phone ?? ""}</p>
                </div>
                <div className="my-1 h-px bg-border" />
                <Link
                  href="/profile"
                  role="menuitem"
                  className="flex items-center gap-2.5 rounded-xl px-3 py-2 text-[13px] font-semibold text-ink transition-colors hover:bg-surface"
                >
                  <UserRound className="h-4 w-4 text-muted" />
                  View profile
                </Link>
                <button
                  role="menuitem"
                  onClick={handleLogout}
                  className="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-[13px] font-semibold text-muted transition-colors hover:bg-surface hover:text-red-500"
                >
                  <LogOut className="h-4 w-4" />
                  Log out
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
