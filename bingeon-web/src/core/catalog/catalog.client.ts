import type { ApiReply } from "~/core/errors/api-error";
import { ApiError, unwrap } from "~/core/errors/api-error";

/**
 * The catalogue: GET /catalog/<action>?<query> on the site's own origin
 * (CloudFront, or `vite dev`'s proxy to it), cached by CloudFront for
 * everyone. The same request must be the same URL, so the query is
 * canonical (see canonicalQuery).
 */

export type CatalogParams = Record<string, string | number | undefined>;

/** A search as the catalogue is asked it: trimmed, single-spaced, lowercased. */
export function canonicalSearch(q: string): string {
  return q.trim().split(/\s+/).join(" ").toLowerCase();
}

/**
 * The query in one form: names sorted, empty values and page 1 left out,
 * the search canonical.
 */
export function canonicalQuery(params: CatalogParams): string {
  const entries = Object.entries(params)
    .map(([name, value]): [string, string] => [name, name === "q" ? canonicalSearch(String(value ?? "")) : String(value ?? "")])
    .filter(([name, value]) => value !== "" && !(name === "page" && value === "1"))
    .sort(([a], [b]) => (a < b ? -1 : 1));
  return new URLSearchParams(entries).toString();
}

export async function catalogGet<T>(action: string, params: CatalogParams = {}): Promise<T> {
  const query = canonicalQuery(params);
  const url = query ? `/catalog/${action}?${query}` : `/catalog/${action}`;
  let response: Response;
  try {
    response = await fetch(url);
  } catch (cause) {
    throw new ApiError("unreachable", "Couldn't reach bingeon. Check the connection and try again.", undefined, cause);
  }
  let reply: ApiReply<T>;
  try {
    reply = (await response.json()) as ApiReply<T>;
  } catch (cause) {
    throw new ApiError("internal", "Something went wrong. Please try again.", undefined, cause);
  }
  return unwrap(reply);
}
