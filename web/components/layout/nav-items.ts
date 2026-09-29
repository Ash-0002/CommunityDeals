import { Home, Tag, User } from "lucide-react";

export const NAV_ITEMS = [
  { href: "/dashboard", label: "Home", icon: Home },
  { href: "/campaigns", label: "Deals", icon: Tag },
  { href: "/profile", label: "Profile", icon: User },
] as const;

/** True when `pathname` belongs to `href`. /campaigns/new still lights up "Deals". */
export function isActive(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(`${href}/`);
}
