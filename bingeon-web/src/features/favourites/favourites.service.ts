import type { Favourite, Status } from "~/features/account/account.types";
import { invoke } from "~/core/lambda/lambda.client";

export const FavouritesService = {
  add: (id: number) => invoke<{ favourite: Favourite; status: Status }>("favourites.add", { id }),
  remove: (id: number) => invoke<{ id: number; status: Status }>("favourites.remove", { id }),
};
