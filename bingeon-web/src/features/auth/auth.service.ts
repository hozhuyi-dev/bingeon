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

  /** Who the API says the caller is: the infra phase's check, until the account exists. */
  whoAmI: () => invoke<{ user: string }>("account.get"),

  signIn: () => signInWithRedirect({ provider: "Google" }),

  signOut: () => signOut(),
};
