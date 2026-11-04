import { useQuery } from "@tanstack/react-query";
import { AccountService } from "./account.service";
import { AuthQuery } from "~/features/auth/auth.query";
import { AccountKeys } from "~/core/constants/query-keys.constant";

export const AccountQuery = {
  /**
   * The signed-in user's account: read once a session, then kept current by
   * each mutation's reply. Nothing while signed out.
   */
  useAccount: () => {
    const auth = AuthQuery.useAuthState();
    return useQuery({
      queryKey: AccountKeys.view(),
      queryFn: () => AccountService.get(),
      enabled: auth.data?.status === "signed-in",
      staleTime: Infinity,
    });
  },
};
