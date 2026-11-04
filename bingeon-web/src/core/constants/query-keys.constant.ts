export const AuthKeys = {
  all: ["auth"] as const,
  state: () => [...AuthKeys.all, "state"] as const,
  identity: () => [...AuthKeys.all, "identity"] as const,
};
