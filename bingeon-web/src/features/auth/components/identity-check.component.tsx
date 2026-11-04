import { AuthMutation } from "../auth.mutation";
import { AuthQuery } from "../auth.query";

/**
 * The infra phase's check: sign in with Google, then show who the api
 * function says the caller is. The web phase replaces it with the header's
 * account menu.
 */
export function IdentityCheck() {
  const auth = AuthQuery.useAuthState();
  const signedIn = auth.data?.status === "signed-in";
  const whoAmI = AuthQuery.useWhoAmI(signedIn);
  const signIn = AuthMutation.useSignIn();
  const signOut = AuthMutation.useSignOut();

  if (auth.isPending) return null;

  return (
    <div className="flex flex-wrap items-center gap-3 text-sm">
      {auth.data?.status === "signed-in" ? (
        <>
          <span>{auth.data.email}</span>
          <span className="text-muted-foreground">
            {whoAmI.isPending ? "Asking the API…" : whoAmI.isError ? whoAmI.error.message : `API user: ${whoAmI.data.user}`}
          </span>
          <button type="button" className="underline" onClick={() => signOut.mutate()}>
            Sign out
          </button>
        </>
      ) : (
        <button type="button" className="underline" onClick={() => signIn.mutate()}>
          Sign in with Google
        </button>
      )}
    </div>
  );
}
