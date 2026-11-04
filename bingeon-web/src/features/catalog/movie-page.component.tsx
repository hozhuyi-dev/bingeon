import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { PlayIcon, StarIcon } from "lucide-react";
import { CatalogQuery } from "./catalog.query";
import { MovieCards } from "./components/movies.component";
import { HeartButton } from "~/features/favourites/components/heart-button.component";
import { Shelf } from "~/components/shared/movies/shelf.component";
import { Button } from "~/components/ui/button";
import { backdropUrl, posterUrl, profileUrl } from "~/core/tmdb/image.util";
import { formatDate, formatRating, formatRuntime } from "~/core/utils/format.util";

export function MoviePage({ id }: { id: number }) {
  const { data: movie } = useSuspenseQuery(CatalogQuery.movie(id));
  const facts = [
    movie.released && formatDate(movie.released),
    movie.runtime ? formatRuntime(movie.runtime) : undefined,
  ].filter(Boolean);

  return (
    <article className="space-y-10">
      <div className="relative -mx-(--page-gutter) overflow-hidden sm:mx-0 sm:rounded-xl">
        {movie.backdrop && (
          <img src={backdropUrl(movie.backdrop)} alt="" className="absolute inset-0 size-full object-cover opacity-30" />
        )}
        <div className="relative flex flex-col gap-6 p-(--page-gutter) sm:flex-row sm:p-8">
          {movie.poster && (
            <img src={posterUrl(movie.poster, "w500")} alt="" className="w-40 shrink-0 self-start rounded-lg shadow-lg sm:w-56" />
          )}
          <div className="space-y-4">
            <div>
              <h1 className="text-2xl font-semibold sm:text-3xl">
                {movie.title} {movie.year ? <span className="font-normal text-muted-foreground">({movie.year})</span> : null}
              </h1>
              <p className="mt-1 flex flex-wrap items-center gap-x-3 text-sm text-muted-foreground">
                {facts.map((fact) => (
                  <span key={fact}>{fact}</span>
                ))}
                {movie.rating ? (
                  <span className="flex items-center gap-1">
                    <StarIcon className="size-3.5" aria-label="Rated" /> {formatRating(movie.rating)}
                  </span>
                ) : null}
              </p>
            </div>
            {movie.genres.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {movie.genres.map((genre) => (
                  <Button key={genre.id} variant="outline" size="xs" render={<Link to="/browse" search={{ genre: genre.id }} />}>
                    {genre.name}
                  </Button>
                ))}
              </div>
            )}
            {movie.tagline && <p className="italic">{movie.tagline}</p>}
            {movie.overview && <p className="max-w-prose leading-relaxed">{movie.overview}</p>}
            {movie.directors.length > 0 && (
              <p className="text-sm">
                <span className="text-muted-foreground">Directed by</span> {movie.directors.join(", ")}
              </p>
            )}
            <div className="flex items-center gap-3">
              <HeartButton id={movie.id} title={movie.title} />
              {movie.trailer && (
                <Button
                  variant="outline"
                  render={<a href={`https://www.youtube.com/watch?v=${movie.trailer}`} target="_blank" rel="noopener noreferrer" />}
                >
                  <PlayIcon /> Watch trailer
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>

      {movie.cast.length > 0 && (
        <Shelf title="Cast">
          {movie.cast.map((person) => (
            <div key={`${person.name}-${person.character}`} className="text-sm">
              <div className="aspect-2/3 overflow-hidden rounded-lg bg-muted">
                {person.photo && <img src={profileUrl(person.photo)} alt="" loading="lazy" className="size-full object-cover" />}
              </div>
              <p className="mt-2 font-medium">{person.name}</p>
              {person.character && <p className="text-xs text-muted-foreground">{person.character}</p>}
            </div>
          ))}
        </Shelf>
      )}
      {movie.recommended.length > 0 && (
        <Shelf title="Recommended">
          <MovieCards movies={movie.recommended} />
        </Shelf>
      )}
      {movie.similar.length > 0 && (
        <Shelf title="Similar">
          <MovieCards movies={movie.similar} />
        </Shelf>
      )}
    </article>
  );
}
