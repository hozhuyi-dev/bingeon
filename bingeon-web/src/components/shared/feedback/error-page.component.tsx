import { useRouter } from "@tanstack/react-router";
import type { ErrorComponentProps } from "@tanstack/react-router";
import { Message } from "~/components/shared/feedback/message.component";
import { Button } from "~/components/ui/button";
import { messageOf } from "~/core/errors/api-error";

/** A page whose data couldn't be read: why, and trying again. */
export function ErrorPage({ error }: ErrorComponentProps) {
  const router = useRouter();
  return (
    <Message
      lost
      title="That didn't work"
      action={
        <Button variant="outline" onClick={() => void router.invalidate()}>
          Try again
        </Button>
      }
    >
      {messageOf(error)}
    </Message>
  );
}
