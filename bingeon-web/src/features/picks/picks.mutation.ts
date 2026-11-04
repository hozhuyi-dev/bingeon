import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { PicksService } from "./picks.service";
import type { AccountView } from "~/features/account/account.types";
import { withPicks } from "~/features/account/account.util";
import { AccountKeys } from "~/core/constants/query-keys.constant";
import { messageOf } from "~/core/errors/api-error";

export const PicksMutation = {
  useMake: () => {
    const queryClient = useQueryClient();
    return useMutation({
      mutationFn: () => PicksService.make(),
      onSuccess: ({ picks, status }) =>
        queryClient.setQueryData<AccountView>(AccountKeys.view(), (view) => view && withPicks(view, picks, status)),
      onError: (error) => toast.error(messageOf(error)),
    });
  },
};
