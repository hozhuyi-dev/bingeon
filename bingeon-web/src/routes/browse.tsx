import { createFileRoute } from "@tanstack/react-router";
import { BrowsePage } from "~/features/catalog/browse-page.component";
import { CatalogQuery } from "~/features/catalog/catalog.query";
import { parseBrowseSearch } from "~/features/catalog/catalog.util";
import { queryClient } from "~/core/config/query-client.config";

export const Route = createFileRoute("/browse")({
  validateSearch: parseBrowseSearch,
  loaderDeps: ({ search }) => search,
  // Home holds the genres' names.
  loader: ({ deps }) =>
    Promise.all([
      queryClient.ensureQueryData(CatalogQuery.list(deps.list, deps.genre, deps.page)),
      queryClient.ensureQueryData(CatalogQuery.home()),
    ]),
  component: function Browse() {
    return <BrowsePage search={Route.useSearch()} />;
  },
});
