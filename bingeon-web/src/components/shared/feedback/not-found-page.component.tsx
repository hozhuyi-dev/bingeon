import { Link } from "@tanstack/react-router";
import { Message } from "~/components/shared/feedback/message.component";
import { Button } from "~/components/ui/button";

export function NotFoundPage() {
  return (
    <Message lost title="Nothing here" action={<Button variant="outline" render={<Link to="/" />}>Go home</Button>}>
      This page flew off, or never was.
    </Message>
  );
}
