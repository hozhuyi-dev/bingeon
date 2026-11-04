import type { AfterSignIn, AuthState } from "./auth.types";
import { loadAuth, mayBeSignedIn } from "~/core/auth/amplify";

/**
 * Signing in leaves the app for Google and comes back to "/"; what to do
 * then is kept in this tab's session storage meanwhile.
 */
const AFTER_SIGN_IN_KEY = "bingeon.afterSignIn";

export const AuthService = {
  async getAuthState(): Promise<AuthState> {
    if (!mayBeSignedIn()) return { status: "signed-out" };
    const { fetchAuthSession } = await loadAuth();
    const session = await fetchAuthSession();
    const idToken = session.tokens?.idToken;
    if (!idToken || !session.credentials) return { status: "signed-out" };
    const email = idToken.payload.email;
    return { status: "signed-in", email: typeof email === "string" ? email : undefined };
  },

  async signIn(after?: AfterSignIn): Promise<void> {
    try {
      if (after) sessionStorage.setItem(AFTER_SIGN_IN_KEY, JSON.stringify(after));
    } catch {
      // Signs in all the same, just without coming back to the page.
    }
    const { signInWithRedirect } = await loadAuth();
    await signInWithRedirect({ provider: "Google" });
  },

  async signOut(): Promise<void> {
    const { signOut } = await loadAuth();
    await signOut();
  },

  /** Deletes the user's sign-in (Cognito's user, with its email). */
  async deleteUser(): Promise<void> {
    const { deleteUser } = await loadAuth();
    await deleteUser();
  },

  /** What to do now that the user is back from signing in, once. */
  takeAfterSignIn(): AfterSignIn | undefined {
    try {
      const stored = sessionStorage.getItem(AFTER_SIGN_IN_KEY);
      sessionStorage.removeItem(AFTER_SIGN_IN_KEY);
      return stored ? (JSON.parse(stored) as AfterSignIn) : undefined;
    } catch {
      return undefined;
    }
  },
};
