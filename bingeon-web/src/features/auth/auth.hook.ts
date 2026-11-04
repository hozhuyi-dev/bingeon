import { AuthService } from "./auth.service";

/** What to do now that the user is back from signing in, once (see AuthService.signIn). */
export function useAfterSignIn() {
  return AuthService.takeAfterSignIn;
}
