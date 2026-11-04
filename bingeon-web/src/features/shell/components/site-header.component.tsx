import { Link } from "@tanstack/react-router";
import { SearchBox } from "./search-box.component";
import { AccountMenu } from "~/features/account/components/account-menu.component";
import { AuthMutation } from "~/features/auth/auth.mutation";
import { AuthQuery } from "~/features/auth/auth.query";
import { Pigeon } from "~/components/shared/brand/pigeon.component";
import { Button } from "~/components/ui/button";

export function SiteHeader() {
  const auth = AuthQuery.useAuthState();
  const signIn = AuthMutation.useSignIn();

  return (
    <header className="sticky top-0 z-10 border-b bg-background/90 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-(--content-width) items-center gap-3 px-(--page-gutter)">
        <Link to="/" className="flex shrink-0 items-center gap-1.5" aria-label="bingeon, home">
          <Pigeon className="size-8" />
          <span className="font-semibold text-brand-strong max-sm:hidden">bingeon</span>
        </Link>
        <div className="flex flex-1 justify-center">
          <SearchBox />
        </div>
        <nav className="flex shrink-0 items-center gap-1">
          <Button variant="ghost" render={<Link to="/me" activeProps={{ className: "bg-muted" }} />}>
            My movies
          </Button>
          {auth.data?.status === "signed-in" ? (
            <AccountMenu email={auth.data.email} />
          ) : (
            <Button variant="outline" onClick={() => signIn.mutate(undefined)} disabled={auth.isPending || signIn.isPending}>
              Sign in
            </Button>
          )}
        </nav>
      </div>
    </header>
  );
}
