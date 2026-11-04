import { createFileRoute } from "@tanstack/react-router";
import { PrivacyPage } from "~/features/privacy/privacy-page.component";

export const Route = createFileRoute("/privacy")({
  component: PrivacyPage,
});
