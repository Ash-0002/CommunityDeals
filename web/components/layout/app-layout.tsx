// Shell for every authenticated page — "community feed" layout.
//
// No sidebar: a floating glass bar sticks to the top of the page (with a gap
// above it so it visibly floats) and, on phones, a floating pill nav sits near
// the bottom. The page itself scrolls naturally instead of an inner pane, so
// mobile browser chrome behaves and long pages feel like a feed.

import { TopNav } from "./top-nav";
import { BottomNav } from "./bottom-nav";

export function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-surface">
      <TopNav />
      <main className="mx-auto w-full max-w-6xl px-4 pb-32 pt-6 sm:px-6 sm:pt-8 md:pb-16">
        {children}
      </main>
      <BottomNav />
    </div>
  );
}
