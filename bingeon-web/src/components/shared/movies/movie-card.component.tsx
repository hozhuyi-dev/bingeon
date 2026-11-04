import { Link } from "@tanstack/react-router";
import { StarIcon } from "lucide-react";
import type { ReactNode } from "react";
import type { MovieSummary } from "./movie.types";
import { posterUrl } from "~/core/tmdb/image.util";
import { formatRating } from "~/core/utils/format.util";

interface MovieCardProps {
  movie: MovieSummary;
  /** On the poster's corner: the heart. */
  action?: ReactNode;
  /** Under the title: why it was picked. */
  children?: ReactNode;
}

export function MovieCard({ movie, action, children }: MovieCardProps) {
  return (
    <article className="group relative min-w-0">
      <Link to="/movie/$id" params={{ id: movie.id }} className="block rounded-lg focus-visible:outline-2 focus-visible:outline-ring">
        <div className="aspect-2/3 overflow-hidden rounded-lg bg-muted">
          {movie.poster ? (
            <img
              src={posterUrl(movie.poster)}
              alt=""
              loading="lazy"
              className="size-full object-cover transition-transform duration-300 group-hover:scale-105"
            />
          ) : (
            <div className="grid size-full place-items-center p-3 text-center text-sm text-muted-foreground">{movie.title}</div>
          )}
        </div>
        <h3 className="mt-2 line-clamp-2 text-sm font-medium">{movie.title}</h3>
        <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
          {movie.year ? <span>{movie.year}</span> : null}
          {movie.rating ? (
            <span className="flex items-center gap-0.5">
              <StarIcon className="size-3" aria-label="Rated" />
              {formatRating(movie.rating)}
            </span>
          ) : null}
        </p>
      </Link>
      {action && <div className="absolute top-2 right-2">{action}</div>}
      {children}
    </article>
  );
}
