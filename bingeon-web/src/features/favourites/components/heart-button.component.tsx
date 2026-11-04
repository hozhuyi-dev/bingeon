import { useLocation } from "@tanstack/react-router";
import { HeartIcon } from "lucide-react";
import { toast } from "sonner";
import { FavouritesMutation } from "../favourites.mutation";
import { AccountQuery } from "~/features/account/account.query";
import { AuthMutation } from "~/features/auth/auth.mutation";
import { AuthQuery } from "~/features/auth/auth.query";
import { cn } from "~/lib/utils";

interface HeartButtonProps {
  id: number;
  title: string;
  className?: string;
}

/**
 * A movie's heart: favourites it, or not. Signed out, it signs in first and
 * favourites the movie on the way back. A change the rules refuse says why
 * (the account's `can`) rather than call the API.
 */
export function HeartButton({ id, title, className }: HeartButtonProps) {
  const auth = AuthQuery.useAuthState();
  const account = AccountQuery.useAccount();
  const add = FavouritesMutation.useAdd();
  const remove = FavouritesMutation.useRemove();
  const signIn = AuthMutation.useSignIn();
  const location = useLocation();

  const signedIn = auth.data?.status === "signed-in";
  const view = account.data;
  const favourite = view?.favourites.some((f) => f.id === id) ?? false;
  const busy = add.isPending || remove.isPending || signIn.isPending || (signedIn && !view);

  function toggle() {
    if (!signedIn) {
      signIn.mutate({ path: location.href, favourite: id });
      return;
    }
    if (!view) return;
    const verdict = favourite ? view.can.remove : view.can.add;
    if (!verdict.ok) {
      toast(verdict.why);
      return;
    }
    if (favourite) remove.mutate(id);
    else add.mutate(id);
  }

  return (
    <button
      type="button"
      onClick={toggle}
      disabled={busy}
      aria-pressed={favourite}
      aria-label={favourite ? `Remove ${title} from favourites` : `Add ${title} to favourites`}
      className={cn(
        "grid size-9 place-items-center rounded-full bg-background/80 shadow-sm backdrop-blur transition hover:scale-110 focus-visible:outline-2 focus-visible:outline-ring disabled:opacity-60",
        className,
      )}
    >
      <HeartIcon className={cn("size-4.5", favourite ? "fill-red-500 text-red-500" : "text-foreground")} />
    </button>
  );
}
