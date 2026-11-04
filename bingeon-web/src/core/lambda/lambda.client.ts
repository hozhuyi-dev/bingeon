import { InvokeCommand, LambdaClient } from "@aws-sdk/client-lambda";
import { fetchAuthSession } from "aws-amplify/auth";
import type { ApiResponse } from "./lambda.types";

/**
 * The signed-in API is one Lambda invoked directly with the user's
 * identity-pool credentials. Every call is an {action, payload} pair and every
 * reply is {data} or {error}.
 */

export class ApiError extends Error {
  readonly code: string;
  /** The invocation's request id, as the Lambda's log shows it. */
  readonly requestId: string | undefined;

  constructor(code: string, message: string, requestId?: string, cause?: unknown) {
    super(message, cause === undefined ? undefined : { cause });
    this.name = "ApiError";
    this.code = code;
    this.requestId = requestId;
  }
}

const client = new LambdaClient({
  region: import.meta.env.VITE_AWS_REGION,
  // Resolved per request so Amplify can refresh expired credentials.
  credentials: async () => {
    const { credentials } = await fetchAuthSession();
    if (!credentials) {
      throw new ApiError("unauthenticated", "You're signed out.");
    }
    return credentials;
  },
});

const encoder = new TextEncoder();
const decoder = new TextDecoder();

export async function invoke<T>(action: string, payload?: unknown): Promise<T> {
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
    const requestId = (cause as { $metadata?: { requestId?: string } } | null)?.$metadata?.requestId;
    if (cause instanceof Error && cause.name === "TooManyRequestsException") {
      throw new ApiError("busy", "bingeon is busy. Try again in a moment.", requestId, cause);
    }
    throw new ApiError("unreachable", "Couldn't reach bingeon. Check the connection and try again.", requestId, cause);
  }
  const requestId = result.$metadata.requestId;

  // A crash or timeout inside the function, as opposed to an error it chose
  // to return.
  if (result.FunctionError || !result.Payload) {
    throw new ApiError("internal", "Something went wrong. Please try again.", requestId);
  }

  const body = JSON.parse(decoder.decode(result.Payload)) as ApiResponse<T>;
  if (body.error) {
    throw new ApiError(body.error.code, body.error.message, requestId);
  }
  // TanStack Query rejects undefined data.
  return (body.data ?? null) as T;
}
