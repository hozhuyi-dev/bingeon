export const AuthKeys = {
  all: ["auth"] as const,
  state: () => [...AuthKeys.all, "state"] as const,
};

export const AccountKeys = {
  all: ["account"] as const,
  view: () => [...AccountKeys.all, "view"] as const,
};

export const CatalogKeys = {
  all: ["catalog"] as const,
  home: () => [...CatalogKeys.all, "home"] as const,
  list: (list: string | undefined, genre: number | undefined, page: number) =>
    [...CatalogKeys.all, "list", { list, genre, page }] as const,
  search: (q: string, page: number) => [...CatalogKeys.all, "search", { q, page }] as const,
  movie: (id: number) => [...CatalogKeys.all, "movie", id] as const,
};
