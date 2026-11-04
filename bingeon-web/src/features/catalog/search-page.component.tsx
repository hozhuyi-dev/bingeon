import { useSuspenseQuery } from "@tanstack/react-query";
import { CatalogQuery } from "./catalog.query";
import { Movies } from "./components/movies.component";
import { Pager } from "./components/pager.component";
import type { SearchSearch } from "./catalog.types";
import { Message } from "~/components/shared/feedback/message.component";

export function SearchPage({ search }: { search: SearchSearch }) {
  if (!search.q) {
    return <Message title="Search for a movie">Type a title in the search box above.</Message>;
  }
  return <Results q={search.q} page={search.page} />;
}

function Results({ q, page }: { q: string; page: number | undefined }) {
  const { data } = useSuspenseQuery(CatalogQuery.search(q, page));
  if (data.results.length === 0) {
    return <Message title={`Nothing found for “${q}”`}>Try another title, or fewer words.</Message>;
  }
  return (
    <div>
      <h1 className="type-page-title mb-6">Results for “{q}”</h1>
      <Movies movies={data.results} />
      <Pager page={data.page} totalPages={data.totalPages} />
    </div>
  );
}
