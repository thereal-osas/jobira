import { Link } from "@tanstack/react-router";
import { cn } from "@/lib/utils";

export function LeafMark({ className }: { className?: string | undefined }) {
  return (
    <svg viewBox="0 0 44 42" aria-hidden="true" className={cn("h-9 w-9", className)}>
      <ellipse cx="14" cy="15" rx="7" ry="13" transform="rotate(-42 14 15)" fill="var(--brand-leaf)" />
      <ellipse cx="29" cy="14" rx="7" ry="12" transform="rotate(38 29 14)" fill="var(--brand-blue)" />
      <ellipse cx="22" cy="29" rx="7" ry="12" fill="var(--brand-leaf-soft)" />
    </svg>
  );
}

export function Logo({ compact = false }: { compact?: boolean }) {
  return (
    <Link to="/" className="flex items-center gap-2.5" aria-label="Jobira home">
      <LeafMark className={compact ? "h-8 w-8" : undefined} />
      <span className="leading-none">
        <span className="block font-display text-[1.45rem] font-semibold text-foreground">Jobira</span>
        {!compact && <span className="mt-1 block text-[0.48rem] font-bold tracking-[0.22em] text-muted-foreground">WORK MEETS OPPORTUNITY</span>}
      </span>
    </Link>
  );
}