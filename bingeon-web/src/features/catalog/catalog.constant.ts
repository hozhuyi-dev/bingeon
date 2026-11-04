import type { Home, ListName } from "./catalog.types";

const HOUR = 60 * 60 * 1000;

/** As long as CloudFront keeps each (bingeon-lambda's catalog/domain.go). */
export const LIST_STALE_MS = 6 * HOUR;
export const SEARCH_STALE_MS = 24 * HOUR;
export const MOVIE_STALE_MS = 24 * HOUR;

/** The lists, as home's shelves show them. */
export const LISTS: ReadonlyArray<{ name: ListName; title: string; shelf: keyof Omit<Home, "genres"> }> = [
  { name: "popular", title: "Popular", shelf: "popular" },
  { name: "now_playing", title: "Now playing in Malaysia", shelf: "nowPlaying" },
  { name: "upcoming", title: "Upcoming in Malaysia", shelf: "upcoming" },
  { name: "top_rated", title: "Top rated", shelf: "topRated" },
];

export const MAX_PAGE = 500;
