import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { PicksPanel } from "./picks-panel.component";
import type { AccountView } from "~/features/account/account.types";

vi.mock("~/features/favourites/components/heart-button.component", () => ({ HeartButton: () => null }));
vi.mock("@tanstack/react-router", () => ({ Link: ({ children }: { children: React.ReactNode }) => <a>{children}</a> }));

const yes = { ok: true };

function account(pick: AccountView["can"]["pick"]): AccountView {
  return {
    can: { add: yes, remove: yes, pick },
    today: { picksLeft: 0, changesLeft: 1000, resets: "2026-11-05T08:00:00Z" },
    favourites: [],
    picks: { movies: [{ id: 1, title: "Thief", year: 1981, reason: "Michael Mann's first." }] },
  };
}

function renderPanel(view: AccountView) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <PicksPanel account={view} />
    </QueryClientProvider>,
  );
}

describe("PicksPanel", () => {
  test("says why it can't make picks, as the API does, and doesn't offer to", () => {
    renderPanel(account({ ok: false, why: "You've made today's 20 picks.", until: "2026-11-05T08:00:00Z" }));
    expect(screen.getByRole("button", { name: /make picks/i })).toBeDisabled();
    expect(screen.getByText("You've made today's 20 picks.")).toBeInTheDocument();
  });

  test("shows each pick with why", () => {
    renderPanel(account(yes));
    expect(screen.getByRole("button", { name: /make picks/i })).toBeEnabled();
    expect(screen.getByRole("heading", { name: "Thief" })).toBeInTheDocument();
    expect(screen.getByText("Michael Mann's first.")).toBeInTheDocument();
  });
});
