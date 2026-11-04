import { fetchAuthSession, signInWithRedirect, signOut } from "aws-amplify/auth";
import type { AuthState } from "./auth.types";
import { invoke } from "~/core/lambda/lambda.client";

export const AuthService = {
  async getAuthState(): Promise<AuthState> {
    const session = await fetchAuthSession();
    const idToken = session.tokens?.idToken;
    if (!idToken || !session.credentials) return { status: "signed-out" };
    const email = idToken.payload.email;
    return { status: "signed-in", email: typeof email === "string" ? email : undefined };
  },

  /** The account as the API has it: the sign-in check's, until the web phase's account page. */
  whoAmI: () => invoke<{ favourites: Array<unknown>; today: { picksLeft: number } }>("account.get"),

  signIn: () => signInWithRedirect({ provider: "Google" }),

  signOut: () => signOut(),
};
