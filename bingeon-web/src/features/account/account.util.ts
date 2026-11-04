import type { AccountView, Favourite, PicksView, Status } from "./account.types";

// Each mutation replies with what it changed and the account's status;
// these put that into the account as it is cached, so nothing is read again.

export function withFavourite(view: AccountView, favourite: Favourite, status: Status): AccountView {
  return { ...view, ...status, favourites: [favourite, ...view.favourites.filter((f) => f.id !== favourite.id)] };
}

export function withoutFavourite(view: AccountView, id: number, status: Status): AccountView {
  return { ...view, ...status, favourites: view.favourites.filter((f) => f.id !== id) };
}

export function withPicks(view: AccountView, picks: PicksView, status: Status): AccountView {
  return { ...view, ...status, picks };
}
