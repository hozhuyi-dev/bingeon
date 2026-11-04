import { useSuspenseQuery } from "@tanstack/react-query";
import { CatalogQuery } from "./catalog.query";
import { LISTS } from "./catalog.constant";
import { GenreChips } from "./components/genre-chips.component";
import { Movies } from "./components/movies.component";
import { Pager } from "./components/pager.component";
import type { BrowseSearch } from "./catalog.types";

/** A page of a list, or of a genre's movies. */
export function BrowsePage({ search }: { search: BrowseSearch }) {
  const { data } = useSuspenseQuery(CatalogQuery.list(search.list, search.genre, search.page));
  // The genre's name is home's, which the catalogue has cached.
  const { data: home } = useSuspenseQuery(CatalogQuery.home());
  const title = search.genre
    ? (home.genres.find((g) => g.id === search.genre)?.name ?? "Genre")
    : (LISTS.find((l) => l.name === search.list)?.title ?? "Movies");

  return (
    <div className="space-y-6">
      <GenreChips genres={home.genres} active={search.genre} />
      <h1 className="type-page-title">{title}</h1>
      <Movies movies={data.results} />
      <Pager page={data.page} totalPages={data.totalPages} />
    </div>
  );
}
