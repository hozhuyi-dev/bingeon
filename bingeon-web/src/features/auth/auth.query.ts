import { useQuery } from "@tanstack/react-query";
import { AuthService } from "./auth.service";
import { AuthKeys } from "~/core/constants/query-keys.constant";

export const AuthQuery = {
  /** Read once a page load: signing in or out leaves the page. */
  useAuthState: () =>
    useQuery({
      queryKey: AuthKeys.state(),
      queryFn: () => AuthService.getAuthState(),
      staleTime: Infinity,
      retry: false,
    }),
};
