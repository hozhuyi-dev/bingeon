import { Link } from "@tanstack/react-router";
import type { Genre } from "../catalog.types";
import { Button } from "~/components/ui/button";

interface GenreChipsProps {
  genres: ReadonlyArray<Genre>;
  /** The genre showing, highlighted. */
  active?: number | undefined;
}

/**
 * The genres, for filtering at a tap: one row that swipes sideways on a
 * phone, wrapping on a wider screen.
 */
export function GenreChips({ genres, active }: GenreChipsProps) {
  return (
    <nav
      aria-label="Genres"
      className="scrollbar-none -mx-(--page-gutter) flex gap-2 overflow-x-auto px-(--page-gutter) md:mx-0 md:flex-wrap md:px-0"
    >
      {genres.map((genre) => (
        <Button
          key={genre.id}
          variant={genre.id === active ? "default" : "outline"}
          size="sm"
          className="shrink-0 rounded-full"
          aria-current={genre.id === active ? "page" : undefined}
          nativeButton={false}
          render={<Link to="/browse" search={{ genre: genre.id }} />}
        >
          {genre.name}
        </Button>
      ))}
    </nav>
  );
}
