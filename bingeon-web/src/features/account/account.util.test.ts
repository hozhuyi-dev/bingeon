import { withFavourite, withoutFavourite } from "./account.util";
import type { AccountView, Status } from "./account.types";

const yes = { ok: true };
const status = (changesLeft: number): Status => ({
  can: { add: yes, remove: yes, pick: yes },
  today: { picksLeft: 20, changesLeft, resets: "2026-11-05T08:00:00Z" },
});
const heat = { id: 949, title: "Heat", added: "2026-11-04T12:00:00Z" };
const view: AccountView = { ...status(1000), favourites: [heat], picks: { movies: [] } };

describe("patching the account", () => {
  test("a new favourite goes first, once, with the new status", () => {
    const thief = { id: 1, title: "Thief", added: "2026-11-04T13:00:00Z" };
    const next = withFavourite(withFavourite(view, thief, status(999)), thief, status(998));
    expect(next.favourites.map((f) => f.id)).toEqual([1, 949]);
    expect(next.today.changesLeft).toBe(998);
  });

  test("a removed one goes", () => {
    expect(withoutFavourite(view, 949, status(999)).favourites).toEqual([]);
  });
});
