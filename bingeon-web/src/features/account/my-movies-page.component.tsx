import { AccountQuery } from "./account.query";
import { AuthMutation } from "~/features/auth/auth.mutation";
import { AuthQuery } from "~/features/auth/auth.query";
import { PicksPanel } from "~/features/picks/components/picks-panel.component";
import { MovieCards } from "~/features/catalog/components/movies.component";
import { Message } from "~/components/shared/feedback/message.component";
import { PageLoading } from "~/components/shared/feedback/page-loading.component";
import { MovieGrid } from "~/components/shared/movies/movie-grid.component";
import { Button } from "~/components/ui/button";
import { messageOf } from "~/core/errors/api-error";

/** The signed-in user's AI picks and favourites. */
export function MyMoviesPage() {
  const auth = AuthQuery.useAuthState();
  const account = AccountQuery.useAccount();
  const signIn = AuthMutation.useSignIn();

  if (auth.isPending) return <PageLoading />;
  if (auth.data?.status !== "signed-in") {
    return (
      <Message
        title="Your movies, in one place"
        action={
          <Button onClick={() => signIn.mutate({ path: "/me" })} disabled={signIn.isPending}>
            Sign in with Google
          </Button>
        }
      >
        Sign in to keep favourites, and get AI picks made from them.
      </Message>
    );
  }
  if (account.isPending) return <PageLoading />;
  if (account.isError) {
    return (
      <Message lost title="That didn't work" action={<Button variant="outline" onClick={() => void account.refetch()}>Try again</Button>}>
        {messageOf(account.error)}
      </Message>
    );
  }

  const { favourites } = account.data;
  return (
    <div className="space-y-12">
      <h1 className="type-page-title">My movies</h1>
      <PicksPanel account={account.data} />
      <section className="space-y-4">
        <h2 className="type-section-title">Favourites</h2>
        {favourites.length > 0 ? (
          <MovieGrid>
            <MovieCards movies={favourites} />
          </MovieGrid>
        ) : (
          <p className="text-sm text-muted-foreground">Tap the heart on any movie to keep it here.</p>
        )}
      </section>
    </div>
  );
}
