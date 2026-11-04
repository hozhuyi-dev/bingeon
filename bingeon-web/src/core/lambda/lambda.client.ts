import type { InvokeCommand as InvokeCommandType, LambdaClient } from "@aws-sdk/client-lambda";
import type { ApiReply } from "~/core/errors/api-error";
import { loadAuth } from "~/core/auth/amplify";
import { ApiError, unwrap } from "~/core/errors/api-error";

/**
 * The signed-in API: one Lambda invoked directly with the user's
 * identity-pool credentials, {action, payload} in, {data} or {error} out.
 * The AWS SDK, like Amplify, loads with the first call.
 */

let loading: Promise<{ client: LambdaClient; InvokeCommand: typeof InvokeCommandType }> | undefined;

function lambda() {
  loading ??= (async () => {
    const { LambdaClient, InvokeCommand } = await import("@aws-sdk/client-lambda");
    const client = new LambdaClient({
      region: import.meta.env.VITE_AWS_REGION,
      // Resolved per call, so Amplify can refresh expired credentials.
      credentials: async () => {
        const { fetchAuthSession } = await loadAuth();
        const { credentials } = await fetchAuthSession();
        if (!credentials) throw new ApiError("unauthenticated", "You're signed out.");
        return credentials;
      },
    });
    return { client, InvokeCommand };
  })();
  return loading;
}

const encoder = new TextEncoder();
const decoder = new TextDecoder();

export async function invoke<T>(action: string, payload?: unknown): Promise<T> {
  const { client, InvokeCommand } = await lambda();
  let result;
  try {
    result = await client.send(
      new InvokeCommand({
        FunctionName: import.meta.env.VITE_API_FUNCTION_NAME,
        Payload: encoder.encode(JSON.stringify({ action, payload })),
      }),
    );
  } catch (cause) {
    if (cause instanceof ApiError) throw cause;
    if (cause instanceof Error && cause.name === "TooManyRequestsException") {
      throw new ApiError("busy", "bingeon is busy. Try again in a moment.", undefined, cause);
    }
    throw new ApiError("unreachable", "Couldn't reach bingeon. Check the connection and try again.", undefined, cause);
  }
  // A crash or timeout inside the function, rather than an error it chose.
  if (result.FunctionError || !result.Payload) {
    throw new ApiError("internal", "Something went wrong. Please try again.");
  }
  return unwrap(JSON.parse(decoder.decode(result.Payload)) as ApiReply<T>);
}
