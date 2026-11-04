#!/usr/bin/env node
// `pnpm secrets:set <environment>`: asks for the TMDB token and Gemini key and
// stores them as SSM SecureStrings at /bingeon/<env>/<name>, where the Lambdas
// read them at a cold start. Leave an answer empty to keep what is stored.
// Typed values aren't echoed, and reach the AWS CLI through a file only this
// script reads, deleted at once, never its arguments. Keep the names in step
// with bingeon-infra/utils/secrets.util.ts.

import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { createInterface } from "node:readline/promises";
import { Writable } from "node:stream";

process.env.ENVIRONMENT = process.argv[2] ?? "";
if (!process.env.ENVIRONMENT) {
  console.error("Usage: pnpm secrets:set <development|production>");
  process.exit(1);
}
const { aws, environment, projectName } = await import("./aws.mjs");

const SECRETS = [
  { name: "tmdb-token", label: "TMDB API read access token" },
  { name: "gemini-key", label: "Gemini API key (from the bingeon Google project, without billing)" },
];

/** Asks without echoing what is typed. */
async function askHidden(question) {
  let muted = false;
  const output = new Writable({
    write(chunk, encoding, done) {
      if (!muted) process.stdout.write(chunk, encoding);
      done();
    },
  });
  const rl = createInterface({ input: process.stdin, output, terminal: true });
  const answer = rl.question(question);
  muted = true;
  try {
    return (await answer).trim();
  } finally {
    rl.close();
    process.stdout.write("\n");
  }
}

const dir = mkdtempSync(join(tmpdir(), `${projectName}-secrets-`));
try {
  for (const { name, label } of SECRETS) {
    const value = await askHidden(`${label} (empty keeps it): `);
    const parameter = `/${projectName}/${environment}/${name}`;
    if (!value) {
      console.log(`  ${parameter}: kept`);
      continue;
    }
    const input = join(dir, `${name}.json`);
    writeFileSync(input, JSON.stringify({ Name: parameter, Value: value, Type: "SecureString", Overwrite: true }), {
      mode: 0o600,
    });
    try {
      aws(["ssm", "put-parameter", "--cli-input-json", `file://${input}`]);
    } finally {
      rmSync(input, { force: true });
    }
    console.log(`  ${parameter}: stored`);
  }
} finally {
  rmSync(dir, { recursive: true, force: true });
}
