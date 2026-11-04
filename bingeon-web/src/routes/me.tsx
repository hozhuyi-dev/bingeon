import { createFileRoute } from "@tanstack/react-router";
import { MyMoviesPage } from "~/features/account/my-movies-page.component";

export const Route = createFileRoute("/me")({
  component: MyMoviesPage,
});
