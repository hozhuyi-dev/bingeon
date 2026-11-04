import type { MovieSummary } from "~/components/shared/movies/movie.types";

// Mirror bingeon-lambda's catalog/domain.go.

export interface MoviePage {
  results: Array<MovieSummary>;
  page: number;
  totalPages: number;
}

export interface Genre {
  id: number;
  name: string;
}

export interface Home {
  popular: Array<MovieSummary>;
  nowPlaying: Array<MovieSummary>;
  upcoming: Array<MovieSummary>;
  topRated: Array<MovieSummary>;
  genres: Array<Genre>;
}

export interface CastMember {
  name: string;
  character?: string;
  photo?: string;
}

export interface Movie extends MovieSummary {
  released?: string;
  runtime?: number;
  tagline?: string;
  overview?: string;
  backdrop?: string;
  genres: Array<Genre>;
  cast: Array<CastMember>;
  directors: Array<string>;
  /** A YouTube video id. */
  trailer?: string;
  similar: Array<MovieSummary>;
  recommended: Array<MovieSummary>;
}

/** What a page of movies is of: a list, or a genre's. */
export interface BrowseSearch {
  list?: ListName | undefined;
  genre?: number | undefined;
  page?: number | undefined;
}

export interface SearchSearch {
  q?: string | undefined;
  page?: number | undefined;
}

export type ListName = "popular" | "now_playing" | "upcoming" | "top_rated";
