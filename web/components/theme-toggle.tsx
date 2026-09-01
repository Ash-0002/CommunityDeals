"use client";

import { useTheme } from "./theme-provider";

export function ThemeToggle() {
  const { theme, toggle } = useTheme();
  return (
    <button
      onClick={toggle}
      title={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
      className="flex h-8 w-8 items-center justify-center rounded-lg text-base text-white/40 transition-colors hover:bg-white/10 hover:text-white/80"
    >
      {theme === "dark" ? "☀️" : "🌙"}
    </button>
  );
}
