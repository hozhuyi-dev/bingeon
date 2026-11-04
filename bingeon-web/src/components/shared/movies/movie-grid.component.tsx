import type { ReactNode } from "react";

/** Posters in rows: two across on a phone, up to six on a wide screen. */
export function MovieGrid({ children }: { children: ReactNode }) {
  return <div className="grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">{children}</div>;
}
