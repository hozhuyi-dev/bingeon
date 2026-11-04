import { useCallback, useEffect, useRef, useState } from "react";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "~/components/ui/button";

interface ShelfProps {
  title: string;
  /** Beside the title: "See all". */
  more?: ReactNode;
  children: ReactNode;
}

/**
 * A titled row of posters that scrolls sideways: swiped on a phone, moved a
 * screenful at a time by its buttons on a wider one. No scrollbar either way.
 */
export function Shelf({ title, more, children }: ShelfProps) {
  const row = useRef<HTMLDivElement>(null);
  const [ends, setEnds] = useState({ start: true, end: true });

  const measure = useCallback(() => {
    const el = row.current;
    if (!el) return;
    setEnds({ start: el.scrollLeft <= 1, end: el.scrollLeft + el.clientWidth >= el.scrollWidth - 1 });
  }, []);

  useEffect(() => {
    const el = row.current;
    if (!el) return;
    measure();
    const resized = new ResizeObserver(measure);
    resized.observe(el);
    return () => resized.disconnect();
  }, [measure]);

  const scroll = (direction: 1 | -1) => {
    const el = row.current;
    // A screenful, less a poster, so the last one seen stays in sight.
    if (el) el.scrollBy({ left: direction * (el.clientWidth - 160), behavior: "smooth" });
  };

  return (
    <section className="min-w-0">
      <div className="flex items-center justify-between gap-4">
        <h2 className="type-section-title">{title}</h2>
        <div className="flex items-center gap-1">
          {more}
          <div className="ml-2 flex gap-1 max-md:hidden">
            <Button variant="outline" size="icon-sm" aria-label={`Scroll ${title} back`} disabled={ends.start} onClick={() => scroll(-1)}>
              <ChevronLeftIcon />
            </Button>
            <Button variant="outline" size="icon-sm" aria-label={`Scroll ${title} on`} disabled={ends.end} onClick={() => scroll(1)}>
              <ChevronRightIcon />
            </Button>
          </div>
        </div>
      </div>
      <div
        ref={row}
        onScroll={measure}
        className="scrollbar-none mt-3 flex snap-x gap-4 overflow-x-auto pb-1 *:w-32 *:shrink-0 *:snap-start sm:*:w-40"
      >
        {children}
      </div>
    </section>
  );
}
