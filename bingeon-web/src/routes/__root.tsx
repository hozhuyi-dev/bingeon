import { Outlet, createRootRoute } from "@tanstack/react-router";
import { IdentityCheck } from "~/features/auth/components/identity-check.component";
import "~/styles/app.css";

export const Route = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  return (
    <>
      <header className="mx-auto flex max-w-(--content-width) justify-end px-(--page-gutter) pt-4">
        <IdentityCheck />
      </header>
      <Outlet />
    </>
  );
}
