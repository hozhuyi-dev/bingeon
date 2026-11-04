import { useNavigate } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { FavouritesMutation } from "../favourites.mutation";
import { AuthQuery } from "~/features/auth/auth.query";
import { useAfterSignIn } from "~/features/auth/auth.hook";

/**
 * Back from signing in (always at "/"): returns to the page the user left,
 * and favourites the movie whose heart they tapped. Renders nothing.
 */
export function AfterSignIn() {
  const auth = AuthQuery.useAuthState();
  const takeAfterSignIn = useAfterSignIn();
  const navigate = useNavigate();
  const add = FavouritesMutation.useAdd();
  const done = useRef(false);

  useEffect(() => {
    if (done.current || auth.data?.status !== "signed-in") return;
    done.current = true;
    const after = takeAfterSignIn();
    if (!after) return;
    void navigate({ href: after.path, replace: true });
    if (after.favourite) add.mutate(after.favourite);
  }, [auth.data, takeAfterSignIn, navigate, add]);

  return null;
}
