import { createFileRoute } from "@tanstack/react-router";
import { HomePage } from "~/features/catalog/home-page.component";
import { CatalogQuery } from "~/features/catalog/catalog.query";
import { queryClient } from "~/core/config/query-client.config";

export const Route = createFileRoute("/")({
  loader: () => queryClient.ensureQueryData(CatalogQuery.home()),
  component: HomePage,
});
