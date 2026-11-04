import { useNavigate, useRouterState } from "@tanstack/react-router";
import { SearchIcon } from "lucide-react";
import { Input } from "~/components/ui/input";
import { canonicalSearch } from "~/core/catalog/catalog.client";

/**
 * Searches when Enter is pressed, not as the user types: each new search
 * is a call no cache has answered yet.
 */
export function SearchBox() {
  const navigate = useNavigate();
  const current = useRouterState({
    select: (s) => (s.location.pathname === "/search" ? String((s.location.search as { q?: unknown }).q ?? "") : ""),
  });

  return (
    <form
      role="search"
      className="relative w-full max-w-sm"
      onSubmit={(event) => {
        event.preventDefault();
        const q = canonicalSearch(String(new FormData(event.currentTarget).get("q") ?? ""));
        if (q) void navigate({ to: "/search", search: { q } });
      }}
    >
      <SearchIcon className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input key={current} name="q" type="search" defaultValue={current} placeholder="Search movies" aria-label="Search movies" className="pl-8" />
    </form>
  );
}
