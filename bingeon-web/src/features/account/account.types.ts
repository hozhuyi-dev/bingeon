import type { MovieSummary } from "~/components/shared/movies/movie.types";

// Mirror bingeon-lambda's views (each slice's domain.go and handler.go).

/** Whether an action may run now, or why not and when that frees up. */
export interface Verdict {
  ok: boolean;
  why?: string;
  until?: string;
}

export interface Status {
  can: { add: Verdict; remove: Verdict; pick: Verdict };
  today: { picksLeft: number; changesLeft: number; resets: string };
}

export interface Favourite extends MovieSummary {
  added: string;
}

export interface Pick extends MovieSummary {
  reason: string;
}

export interface PicksView {
  movies: Array<Pick>;
  madeAt?: string;
}

export interface AccountView extends Status {
  favourites: Array<Favourite>;
  picks: PicksView;
}
