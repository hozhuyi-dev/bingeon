#!/usr/bin/env node
import * as assert from "assert";
import * as cdk from "aws-cdk-lib/core";
import * as dotenv from "dotenv";
import { ApiStack } from "../lib/api-stack";
import { AuthStack } from "../lib/auth-stack";
import { DataStack } from "../lib/data-stack";
import { WebStack } from "../lib/web-stack";
import { appUrls } from "../utils/app-urls.util";
import { resolveEnvironment } from "../utils/environment.util";
import { parseEnvList } from "../utils/parsers.util";

const app = new cdk.App();

// Resolved before dotenv runs: it selects which .env.<environment> to load.
const environment = resolveEnvironment();
dotenv.config({ path: "./.env", quiet: true });
dotenv.config({ path: `./.env.${environment}`, quiet: true });

const projectName = process.env.PROJECT_NAME;
const awsAccountId = process.env.AWS_ACCOUNT_ID;
const awsRegion = process.env.AWS_REGION;
const callbackUrls = parseEnvList(process.env.CALLBACK_URLS);
const logoutUrls = parseEnvList(process.env.LOGOUT_URLS);
const googleClientId = process.env.GOOGLE_CLIENT_ID;
const googleClientSecret = process.env.GOOGLE_CLIENT_SECRET;

assert.ok(projectName, "PROJECT_NAME is required");
assert.ok(awsAccountId, "AWS_ACCOUNT_ID is required");
assert.ok(awsRegion, "AWS_REGION is required");
// Production signs in only at its own site's URL; development also at localhost.
assert.ok(environment !== "development" || callbackUrls.length > 0, "CALLBACK_URLS are required");
assert.ok(googleClientId, "GOOGLE_CLIENT_ID is required");
assert.ok(googleClientSecret, "GOOGLE_CLIENT_SECRET is required");

const env = { account: awsAccountId, region: awsRegion };

// The site comes first: sign-in needs its URL, which exists only once
// CloudFront has made it.
const webStack = new WebStack(app, `${projectName}-${environment}-web`, {
  env: env,
  projectName: projectName,
  environment: environment,
});

const urls = appUrls(webStack.siteOrigin, { callbackUrls, logoutUrls });

const authStack = new AuthStack(app, `${projectName}-${environment}-auth`, {
  env: env,
  projectName: projectName,
  environment: environment,
  callbackUrls: urls.callbackUrls,
  logoutUrls: urls.logoutUrls,
  googleClientId: googleClientId,
  googleClientSecret: googleClientSecret,
});

const dataStack = new DataStack(app, `${projectName}-${environment}-data`, {
  env: env,
  projectName: projectName,
  environment: environment,
});

const apiStack = new ApiStack(app, `${projectName}-${environment}-api`, {
  env: env,
  projectName: projectName,
  environment: environment,
  table: dataStack.data.table,
  authenticatedRole: authStack.auth.authenticatedRole,
});

for (const stack of [webStack, authStack, dataStack, apiStack]) {
  cdk.Tags.of(stack).add("project", projectName);
  cdk.Tags.of(stack).add("environment", environment);
  cdk.Tags.of(stack).add("managed-by", "cdk");
}
