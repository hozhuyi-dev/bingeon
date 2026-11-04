import { useQuery } from "@tanstack/react-query";
import { AuthService } from "./auth.service";
import { AuthKeys } from "~/core/constants/query-keys.constant";

export const AuthQuery = {
  useAuthState: () =>
    useQuery({
      queryKey: AuthKeys.state(),
      queryFn: () => AuthService.getAuthState(),
      staleTime: Infinity,
    }),

  useWhoAmI: (enabled: boolean) =>
    useQuery({
      queryKey: AuthKeys.identity(),
      queryFn: () => AuthService.whoAmI(),
      enabled: enabled,
      staleTime: Infinity,
      retry: false,
    }),
};
