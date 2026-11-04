import { MovieGrid } from "~/components/shared/movies/movie-grid.component";

/** A page still loading: the shape of a grid of posters. */
export function PageLoading() {
  return (
    <div aria-busy="true" aria-label="Loading">
      <div className="mb-6 h-6 w-48 animate-pulse rounded bg-muted" />
      <MovieGrid>
        {Array.from({ length: 12 }, (_, i) => (
          <div key={i} className="aspect-2/3 animate-pulse rounded-lg bg-muted" />
        ))}
      </MovieGrid>
    </div>
  );
}
