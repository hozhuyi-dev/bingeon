import { SparklesIcon } from "lucide-react";
import { PicksMutation } from "../picks.mutation";
import type { AccountView } from "~/features/account/account.types";
import { HeartButton } from "~/features/favourites/components/heart-button.component";
import { MovieCard } from "~/components/shared/movies/movie-card.component";
import { MovieGrid } from "~/components/shared/movies/movie-grid.component";
import { Button } from "~/components/ui/button";
import { formatTime } from "~/core/utils/format.util";

/** AI picks: the last ones, each with why, and making more. */
export function PicksPanel({ account }: { account: AccountView }) {
  const make = PicksMutation.useMake();
  const verdict = account.can.pick;
  const { picksLeft, resets } = account.today;

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="type-section-title">AI picks</h2>
          <p className="text-sm text-muted-foreground">
            {picksLeft} left today · more from {formatTime(new Date(resets))}
          </p>
        </div>
        <Button onClick={() => make.mutate()} disabled={!verdict.ok || make.isPending}>
          <SparklesIcon />
          {make.isPending ? "Asking Gemini…" : "Make picks"}
        </Button>
      </div>
      {!verdict.ok && <p className="text-sm text-muted-foreground">{verdict.why}</p>}
      {account.picks.movies.length > 0 ? (
        <MovieGrid>
          {account.picks.movies.map((pick) => (
            <MovieCard key={pick.id} movie={pick} action={<HeartButton id={pick.id} title={pick.title} />}>
              <p className="mt-1 text-xs text-muted-foreground">{pick.reason}</p>
            </MovieCard>
          ))}
        </MovieGrid>
      ) : (
        <p className="text-sm text-muted-foreground">Picks are movies like your favourites, chosen by Gemini.</p>
      )}
    </section>
  );
}
