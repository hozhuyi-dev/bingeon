import { createFileRoute } from "@tanstack/react-router";
import { HomePage } from "~/features/catalog/home-page.component";

export const Route = createFileRoute("/")({
  component: HomePage,
});
