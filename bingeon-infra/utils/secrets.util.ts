/**
 * The keys the Lambdas read at a cold start, as SSM SecureStrings that
 * `pnpm secrets:set <env>` writes (CloudFormation can't create
 * SecureStrings, so they aren't stack resources). Keep in step with
 * scripts/secrets.mjs.
 */
export const SECRETS = ["tmdb-token", "gemini-key"] as const;

export type Secret = (typeof SECRETS)[number];

/** "/bingeon/development/tmdb-token" */
export function secretParameter(projectName: string, environment: string, secret: Secret): string {
  return `/${projectName}/${environment}/${secret}`;
}

/** "tmdb-token" → "TMDB_TOKEN_PARAM", the variable that names its parameter. */
export function secretVariable(secret: Secret): string {
  return `${secret.toUpperCase().replaceAll("-", "_")}_PARAM`;
}
