import { createFileRoute, notFound } from "@tanstack/react-router";
import { MoviePage } from "~/features/catalog/movie-page.component";
import { CatalogQuery } from "~/features/catalog/catalog.query";
import { queryClient } from "~/core/config/query-client.config";
import { ApiError } from "~/core/errors/api-error";

export const Route = createFileRoute("/movie/$id")({
  params: {
    parse: ({ id }) => ({ id: Number(id) }),
    stringify: ({ id }) => ({ id: String(id) }),
  },
  loader: async ({ params }) => {
    if (!Number.isInteger(params.id) || params.id <= 0) throw notFound();
    try {
      await queryClient.ensureQueryData(CatalogQuery.movie(params.id));
    } catch (error) {
      if (error instanceof ApiError && error.code === "not_found") throw notFound();
      throw error;
    }
  },
  component: function Movie() {
    return <MoviePage id={Route.useParams().id} />;
  },
});
