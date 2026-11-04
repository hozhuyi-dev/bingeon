import type { ReactNode } from "react";
import { Pigeon } from "~/components/shared/brand/pigeon.component";

interface MessageProps {
  title: string;
  children?: ReactNode;
  /** An error, or a page that isn't there. */
  lost?: boolean;
  /** What to do about it: a button or link. */
  action?: ReactNode;
}

/** The pigeon, and a line or two: what an empty list, error or missing page shows. */
export function Message({ title, children, lost, action }: MessageProps) {
  return (
    <div className="mx-auto flex max-w-sm flex-col items-center gap-3 py-16 text-center">
      <Pigeon lost={lost} className="size-20" />
      <h2 className="type-section-title">{title}</h2>
      {children && <div className="text-sm text-muted-foreground">{children}</div>}
      {action}
    </div>
  );
}
