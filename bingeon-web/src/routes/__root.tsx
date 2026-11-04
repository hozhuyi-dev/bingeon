import { Outlet, createRootRoute } from "@tanstack/react-router";
import { SiteHeader } from "~/features/shell/components/site-header.component";
import { SiteFooter } from "~/features/shell/components/site-footer.component";
import { AfterSignIn } from "~/features/favourites/components/after-sign-in.component";
import { Toaster } from "~/components/ui/sonner";
import "~/styles/app.css";

export const Route = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-(--content-width) flex-1 px-(--page-gutter) py-8">
        <Outlet />
      </main>
      <SiteFooter />
      <AfterSignIn />
      <Toaster position="bottom-center" />
    </div>
  );
}
