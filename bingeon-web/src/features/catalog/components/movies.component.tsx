import type { MovieSummary } from "~/components/shared/movies/movie.types";
import { HeartButton } from "~/features/favourites/components/heart-button.component";
import { MovieCard } from "~/components/shared/movies/movie-card.component";
import { MovieGrid } from "~/components/shared/movies/movie-grid.component";

/** Movies as cards, each with its heart: in a grid, or as a shelf's row. */
export function MovieCards({ movies }: { movies: ReadonlyArray<MovieSummary> }) {
  return movies.map((movie) => <MovieCard key={movie.id} movie={movie} action={<HeartButton id={movie.id} title={movie.title} />} />);
}

export function Movies({ movies }: { movies: ReadonlyArray<MovieSummary> }) {
  return (
    <MovieGrid>
      <MovieCards movies={movies} />
    </MovieGrid>
  );
}
