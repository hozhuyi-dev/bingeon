import { cn } from "~/lib/utils";

interface PigeonProps {
  className?: string | undefined;
  /** Lost: a question over its head, for errors and pages that aren't there. */
  lost?: boolean | undefined;
}

/** bingeon's pigeon, its neck in the brand's sheen. Decorative. */
export function Pigeon({ className, lost }: PigeonProps) {
  return (
    <svg viewBox="0 0 64 64" aria-hidden="true" className={cn("size-16", className)}>
      <path d="M15 41 L3 45 L16 49 Z" className="fill-muted-foreground/45" />
      <ellipse cx="32" cy="42" rx="19" ry="12" className="fill-muted-foreground/35" />
      <path d="M20 39 Q34 31 47 40 Q35 47 20 39 Z" className="fill-muted-foreground/55" />
      <path d="M41 29 Q46 35 45 41 L51 39 Q52 32 49 26 Z" className="fill-brand" />
      <circle cx="47" cy="22" r="8" className="fill-muted-foreground/45" />
      <circle cx="49.5" cy="20.5" r="1.6" className="fill-foreground" />
      <path d="M54 21 L60 23 L54 25 Z" className="fill-foreground/70" />
      <path d="M29 53 L28 60 M36 53 L37 60" strokeWidth="2" strokeLinecap="round" className="stroke-muted-foreground/60" />
      {lost && (
        <text x="38" y="9" fontSize="12" fontWeight="700" className="fill-brand">
          ?
        </text>
      )}
    </svg>
  );
}
