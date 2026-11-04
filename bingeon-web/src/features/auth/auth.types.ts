export type AuthState = { status: "signed-out" } | { status: "signed-in"; email: string | undefined };

/** Where to come back to after signing in, and the movie to favourite then. */
export interface AfterSignIn {
  path: string;
  favourite?: number;
}
