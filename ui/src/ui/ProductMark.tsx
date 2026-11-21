/**
 * ProductMark — the Stem product mark: the "ST" tile that answers "which tool
 * am I in". Shared by the rail lockup and the login dialog so the two cannot
 * drift. Decorative: the product name always renders beside it.
 */
interface ProductMarkProps {
  /** Tailwind size + radius + text classes; defaults to the rail's 36px tile. */
  className?: string;
}

export function ProductMark({ className = 'h-9 w-9 rounded-[11px] text-sm' }: ProductMarkProps) {
  return (
    <div
      aria-hidden="true"
      className={`bg-brand-primary flex-center shrink-0 ${className}`}
      data-testid="product-mark"
    >
      {/* allow-hardcoded: the monogram is part of the mark, not copy. */}
      <span className="figure font-extrabold tracking-tight text-on-brand">ST</span>
    </div>
  );
}
