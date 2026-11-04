import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { FavouritesService } from "./favourites.service";
import type { AccountView } from "~/features/account/account.types";
import { withFavourite, withoutFavourite } from "~/features/account/account.util";
import { AccountKeys } from "~/core/constants/query-keys.constant";
import { messageOf } from "~/core/errors/api-error";

export const FavouritesMutation = {
  useAdd: () => {
    const queryClient = useQueryClient();
    return useMutation({
      mutationFn: (id: number) => FavouritesService.add(id),
      onSuccess: ({ favourite, status }) =>
        queryClient.setQueryData<AccountView>(AccountKeys.view(), (view) => view && withFavourite(view, favourite, status)),
      onError: (error) => toast.error(messageOf(error)),
    });
  },
  useRemove: () => {
    const queryClient = useQueryClient();
    return useMutation({
      mutationFn: (id: number) => FavouritesService.remove(id),
      onSuccess: ({ id, status }) =>
        queryClient.setQueryData<AccountView>(AccountKeys.view(), (view) => view && withoutFavourite(view, id, status)),
      onError: (error) => toast.error(messageOf(error)),
    });
  },
};
