import { render, screen } from "@testing-library/react";
import { HomePage } from "./home-page.component";

test("names the app", () => {
  render(<HomePage />);
  expect(screen.getByRole("heading", { name: "bingeon" })).toBeInTheDocument();
});
