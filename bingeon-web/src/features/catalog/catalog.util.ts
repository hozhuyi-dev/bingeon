import { LISTS, MAX_PAGE } from "./catalog.constant";
import type { BrowseSearch, ListName, SearchSearch } from "./catalog.types";

function pageOf(value: unknown): number | undefined {
  const page = Number(value);
  return Number.isInteger(page) && page > 1 && page <= MAX_PAGE ? page : undefined;
}

/** /browse's search params as the URL has them: a list or a genre, popular when neither. */
export function parseBrowseSearch(search: Record<string, unknown>): BrowseSearch {
  const genre = Number(search.genre);
  if (Number.isInteger(genre) && genre > 0) return { genre, page: pageOf(search.page) };
  const list = LISTS.find((l) => l.name === search.list)?.name ?? ("popular" satisfies ListName);
  return { list, page: pageOf(search.page) };
}

/** /search's search params as the URL has them. */
export function parseSearchSearch(search: Record<string, unknown>): SearchSearch {
  const q = typeof search.q === "string" ? search.q : undefined;
  return { q: q?.trim() ? q : undefined, page: pageOf(search.page) };
}
