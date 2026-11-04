// What the scripts share: which deployed environment they act on, and the
// AWS CLI pointed at it.
//
// Locally: reads PROJECT_NAME / AWS_REGION from bingeon-infra/.env[.<env>] and
// uses the AWS_PROFILE profile (default hozhuyi.dev).
// In CI: reads them from the environment and uses the ambient credentials.

import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { parseEnv } from "node:util";

const ENVIRONMENTS = ["development", "production"];

export const root = join(import.meta.dirname, "..");
export const environment = process.env.ENVIRONMENT?.trim() || "development";
const isCI = Boolean(process.env.CI);

if (!ENVIRONMENTS.includes(environment)) {
  console.error(`Unknown environment "${environment}". Expected one of: ${ENVIRONMENTS.join(", ")}.`);
  process.exit(1);
}

function readEnvFile(path) {
  return existsSync(path) ? parseEnv(readFileSync(path, "utf8")) : {};
}

const infraEnv = {
  ...readEnvFile(join(root, "bingeon-infra", ".env")),
  ...readEnvFile(join(root, "bingeon-infra", `.env.${environment}`)),
};

export const projectName = process.env.PROJECT_NAME || infraEnv.PROJECT_NAME || "bingeon";
export const region = process.env.AWS_REGION || infraEnv.AWS_REGION;
export const profile = isCI ? undefined : process.env.AWS_PROFILE || "hozhuyi.dev";

if (!region) {
  console.error(`AWS_REGION is not set (looked in bingeon-infra/.env.${environment}).`);
  process.exit(1);
}

/**
 * Runs the AWS CLI against the environment's region and profile; returns its
 * stdout, or throws with what it said on stderr.
 */
export function aws(args) {
  try {
    return execFileSync("aws", [...args, "--region", region, ...(profile ? ["--profile", profile] : [])], {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    });
  } catch (error) {
    throw new Error(error.stderr?.trim() || error.message);
  }
}
