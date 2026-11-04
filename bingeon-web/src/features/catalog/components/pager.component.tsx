import { Link } from "@tanstack/react-router";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { Button } from "~/components/ui/button";

interface PagerProps {
  page: number;
  totalPages: number;
}

/** Previous and Next, the page in the URL, so Back and a link both work. */
export function Pager({ page, totalPages }: PagerProps) {
  if (totalPages <= 1) return null;
  const to = (n: number) => (n === 1 ? undefined : n);
  return (
    <nav aria-label="Pages" className="mt-8 flex items-center justify-center gap-3 text-sm">
      <Button
        variant="outline"
        disabled={page <= 1}
        nativeButton={page <= 1}
        render={page > 1 ? <Link to="." search={(prev) => ({ ...prev, page: to(page - 1) })} /> : undefined}
      >
        <ChevronLeftIcon /> Previous
      </Button>
      <span className="text-muted-foreground">
        Page {page} of {totalPages}
      </span>
      <Button
        variant="outline"
        disabled={page >= totalPages}
        nativeButton={page >= totalPages}
        render={page < totalPages ? <Link to="." search={(prev) => ({ ...prev, page: page + 1 })} /> : undefined}
      >
        Next <ChevronRightIcon />
      </Button>
    </nav>
  );
}
