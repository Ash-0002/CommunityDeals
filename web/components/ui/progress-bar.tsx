export function ProgressBar({ value, height = 8 }: { value: number; height?: number }) {
  const pct = Math.min(100, Math.max(0, Math.round(value * 100)));
  return (
    <div className="overflow-hidden rounded-full bg-surface" style={{ height }}>
      <div className="h-full rounded-full bg-primary-600 transition-all" style={{ width: `${pct}%` }} />
    </div>
  );
}
