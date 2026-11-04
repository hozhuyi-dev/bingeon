import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
// Finishes a sign-in redirect even though the code that started it is
// split into a lazily loaded chunk.
import "aws-amplify/auth/enable-oauth-listener";
import { getRouter } from "./router";
import { configureAmplify } from "./core/config/amplify.config";
import { queryClient } from "./core/config/query-client.config";

configureAmplify();
const router = getRouter();

const rootElement = document.getElementById("app");
if (!rootElement) {
  throw new Error("Root element #app not found");
}

createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
);
