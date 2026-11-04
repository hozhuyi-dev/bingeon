import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { CatalogQuery } from "./catalog.query";
import { LISTS } from "./catalog.constant";
import { GenreChips } from "./components/genre-chips.component";
import { MovieCards } from "./components/movies.component";
import { Shelf } from "~/components/shared/movies/shelf.component";

export function HomePage() {
  const { data: home } = useSuspenseQuery(CatalogQuery.home());

  return (
    <div className="space-y-10">
      <GenreChips genres={home.genres} />
      {LISTS.map((list) => (
        <Shelf
          key={list.name}
          title={list.title}
          more={
            <Link to="/browse" search={{ list: list.name }} className="text-sm text-muted-foreground hover:text-foreground">
              See all
            </Link>
          }
        >
          <MovieCards movies={home[list.shelf]} />
        </Shelf>
      ))}
    </div>
  );
}
