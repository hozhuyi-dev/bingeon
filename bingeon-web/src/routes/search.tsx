import { createFileRoute } from "@tanstack/react-router";
import { SearchPage } from "~/features/catalog/search-page.component";
import { CatalogQuery } from "~/features/catalog/catalog.query";
import { parseSearchSearch } from "~/features/catalog/catalog.util";
import { queryClient } from "~/core/config/query-client.config";

export const Route = createFileRoute("/search")({
  validateSearch: parseSearchSearch,
  loaderDeps: ({ search }) => search,
  loader: ({ deps }) => (deps.q ? queryClient.ensureQueryData(CatalogQuery.search(deps.q, deps.page)) : undefined),
  component: function Search() {
    return <SearchPage search={Route.useSearch()} />;
  },
});
