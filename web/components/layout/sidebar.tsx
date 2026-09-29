"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { LayoutDashboard, Tag, LogOut } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { useUser, userInitials } from "@/lib/use-user";
import { getRefreshToken, clearTokens } from "@/lib/auth";
import { api } from "@/lib/api";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/campaigns", label: "Campaigns", icon: Tag },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useUser();

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
    <aside className="flex w-56 flex-shrink-0 flex-col border-r border-border bg-card px-3 py-5">
      {/* Logo */}
      <Link href="/dashboard" className="mb-6 flex items-center gap-2.5 px-2">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary-600 text-xs font-bold text-white">
          C
        </div>
        <span className="text-[15px] font-bold text-ink">CommunityDeals</span>
      </Link>

      {/* Nav links */}
      <nav className="flex flex-col gap-0.5">
        {navItems.map((link) => {
          const active = pathname.startsWith(link.href);
          const Icon = link.icon;
          return (
            <Link
              key={link.href}
              href={link.href}
              className={`flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] font-medium transition-colors ${
                active ? "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300" : "text-muted hover:bg-surface hover:text-ink"
              }`}
            >
              <Icon className="h-4 w-4" strokeWidth={2} />
              {link.label}
            </Link>
          );
        })}
      </nav>

      {/* Bottom */}
      <div className="mt-auto border-t border-border pt-3">
        <div className="mb-1 flex items-center justify-between px-2">
          <span className="text-[11px] font-medium text-muted">Appearance</span>
          <ThemeToggle />
        </div>

        <Link
          href="/profile"
          className={`flex items-center gap-2.5 rounded-lg px-2 py-2 transition-colors ${
            pathname.startsWith("/profile") ? "bg-primary-50 dark:bg-primary-900/30" : "hover:bg-surface"
          }`}
        >
          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary-600 text-xs font-bold text-white">
            {userInitials(user?.name)}
          </div>
          <div className="min-w-0">
            <p className="truncate text-[12px] font-semibold text-ink">{user?.name || "My profile"}</p>
            <p className="truncate text-[11px] text-muted">{user?.phone ?? ""}</p>
          </div>
        </Link>

        <button
          onClick={handleLogout}
          className="mt-1 flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] font-medium text-muted transition-colors hover:bg-surface hover:text-red-500"
        >
          <LogOut className="h-4 w-4" strokeWidth={2} />
          Log out
        </button>
      </div>
    </aside>
  );
}
