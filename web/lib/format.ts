// Formatting helpers. Prices from the backend are integers in paise.

export function formatPaise(paise: number): string {
  return `₹${Math.round(paise / 100).toLocaleString("en-IN")}`;
}

export function formatDate(iso: string | null | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString("en-IN", { day: "numeric", month: "short", year: "numeric" });
}

export function formatDateLong(iso: string | null | undefined): string {
  if (!iso) return "Date TBD";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "Date TBD";
  return d.toLocaleDateString("en-IN", { weekday: "long", day: "numeric", month: "long" });
}

export function formatTime(iso: string | null | undefined): string {
  if (!iso) return "Time TBD";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "Time TBD";
  return d.toLocaleTimeString("en-IN", { hour: "2-digit", minute: "2-digit", hour12: true });
}

type Located = {
  city?: string;
  state?: string;
  pin_code?: string;
  address?: string;
};

/** Short "City, State" line — e.g. "Mumbai, Maharashtra". Empty when unknown. */
export function shortLocation(c: Located | null | undefined): string {
  if (!c) return "";
  return [c.city, c.state].filter(Boolean).join(", ");
}

/** Full service address — e.g. "12 Palm Rd, Mumbai, Maharashtra 400053". */
export function fullLocation(c: Located | null | undefined): string {
  if (!c) return "";
  const line = [c.address, c.city, c.state].filter(Boolean).join(", ");
  if (!line) return c.pin_code ?? "";
  return c.pin_code ? `${line} ${c.pin_code}` : line;
}

export function daysLeft(iso: string | null | undefined): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const diff = Math.floor((d.getTime() - Date.now()) / 86_400_000);
  if (diff < 0) return "Ended";
  if (diff === 0) return "Ends today";
  return `${diff} ${diff === 1 ? "day" : "days"} left`;
}
