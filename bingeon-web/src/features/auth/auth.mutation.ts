import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { AuthService } from "./auth.service";
import type { AfterSignIn } from "./auth.types";
import { messageOf } from "~/core/errors/api-error";

/** Both leave the app for Google or Cognito and come back signed in or out. */
export const AuthMutation = {
  useSignIn: () =>
    useMutation({
      mutationFn: (after?: AfterSignIn) => AuthService.signIn(after),
      onError: (error) => toast.error(messageOf(error)),
    }),
  useSignOut: () =>
    useMutation({
      mutationFn: () => AuthService.signOut(),
      onError: (error) => toast.error(messageOf(error)),
    }),
};
