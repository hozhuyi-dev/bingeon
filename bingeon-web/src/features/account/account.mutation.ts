import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { AccountService } from "./account.service";
import { AuthService } from "~/features/auth/auth.service";
import { messageOf } from "~/core/errors/api-error";

export const AccountMutation = {
  /**
   * Deletes everything bingeon keeps of the user, then their sign-in, which
   * signs them out; the page starts over, signed out.
   */
  useDeleteAccount: () =>
    useMutation({
      mutationFn: async () => {
        await AccountService.delete();
        await AuthService.deleteUser();
      },
      onSuccess: () => window.location.assign("/"),
      onError: (error) => toast.error(messageOf(error)),
    }),
};
