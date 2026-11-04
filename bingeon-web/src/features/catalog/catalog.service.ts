import type { Home, ListName, Movie, MoviePage } from "./catalog.types";
import { catalogGet } from "~/core/catalog/catalog.client";

export const CatalogService = {
  home: () => catalogGet<Home>("catalog.home"),
  list: (list: ListName | undefined, genre: number | undefined, page: number) =>
    catalogGet<MoviePage>("movies.list", { list, genre, page }),
  search: (q: string, page: number) => catalogGet<MoviePage>("movies.search", { q, page }),
  movie: (id: number) => catalogGet<Movie>("movies.get", { id }),
};
