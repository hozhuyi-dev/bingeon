import { createRouter } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";
import { ErrorPage } from "~/components/shared/feedback/error-page.component";
import { NotFoundPage } from "~/components/shared/feedback/not-found-page.component";
import { PageLoading } from "~/components/shared/feedback/page-loading.component";

export function getRouter() {
  return createRouter({
    routeTree,
    // A link hovered starts its page's loader: the catalogue is cached, so
    // it costs nothing twice.
    defaultPreload: "intent",
    defaultPendingComponent: PageLoading,
    defaultErrorComponent: ErrorPage,
    defaultNotFoundComponent: NotFoundPage,
    scrollRestoration: true,
  });
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>;
  }
}
