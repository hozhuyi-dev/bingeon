import { queryOptions } from "@tanstack/react-query";
import { CatalogService } from "./catalog.service";
import { LIST_STALE_MS, MOVIE_STALE_MS, SEARCH_STALE_MS } from "./catalog.constant";
import type { ListName } from "./catalog.types";
import { canonicalSearch } from "~/core/catalog/catalog.client";
import { CatalogKeys } from "~/core/constants/query-keys.constant";

/**
 * The catalogue's queries, which a route's loader reads before its page
 * shows (so nothing jumps) and the page then reads again from the cache.
 * Each stays fresh as long as CloudFront keeps it.
 */
export const CatalogQuery = {
  home: () => queryOptions({ queryKey: CatalogKeys.home(), queryFn: () => CatalogService.home(), staleTime: LIST_STALE_MS }),

  list: (list: ListName | undefined, genre: number | undefined, page = 1) =>
    queryOptions({
      queryKey: CatalogKeys.list(list, genre, page),
      queryFn: () => CatalogService.list(list, genre, page),
      staleTime: LIST_STALE_MS,
    }),

  search: (q: string, page = 1) =>
    queryOptions({
      queryKey: CatalogKeys.search(canonicalSearch(q), page),
      queryFn: () => CatalogService.search(q, page),
      staleTime: SEARCH_STALE_MS,
    }),

  movie: (id: number) =>
    queryOptions({ queryKey: CatalogKeys.movie(id), queryFn: () => CatalogService.movie(id), staleTime: MOVIE_STALE_MS }),
};
