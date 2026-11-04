export type AuthState = { status: "signed-out" } | { status: "signed-in"; email: string | undefined };
