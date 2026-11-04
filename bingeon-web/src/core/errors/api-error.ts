/**
 * An error either API chose to return ({error} in the reply), or a call
 * that never got one: code says which (see each client).
 */
export class ApiError extends Error {
  readonly code: string;
  /** When a refusal frees up. */
  readonly until: Date | undefined;

  constructor(code: string, message: string, until?: string, cause?: unknown) {
    super(message, cause === undefined ? undefined : { cause });
    this.name = "ApiError";
    this.code = code;
    this.until = until ? new Date(until) : undefined;
  }
}

/** Every reply from either API: the data, or an error it chose to return. */
export interface ApiReply<T> {
  data?: T;
  error?: { code: string; message: string; until?: string };
}

/** The reply's data, or its error thrown. */
export function unwrap<T>(reply: ApiReply<T>): T {
  if (reply.error) throw new ApiError(reply.error.code, reply.error.message, reply.error.until);
  // An empty result is left out; queries need null, not undefined.
  return (reply.data ?? null) as T;
}

/** What to tell the user about any error. */
export function messageOf(error: unknown): string {
  return error instanceof ApiError ? error.message : "Something went wrong. Please try again.";
}
