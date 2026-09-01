import { Sidebar } from "./sidebar";

// Wraps any authenticated page with the dark sidebar.
// Usage: wrap your page component's return value with <AppLayout>.
export function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen overflow-hidden bg-surface">
      <Sidebar />
      <main className="flex flex-1 flex-col overflow-hidden">{children}</main>
    </div>
  );
}
