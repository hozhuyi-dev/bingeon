#!/usr/bin/env node
import * as assert from "assert";
import * as cdk from "aws-cdk-lib/core";
import * as dotenv from "dotenv";
import { resolveEnvironment } from "../utils/environment.util";

// Resolved before dotenv runs: it selects which .env.<environment> to load.
const environment = resolveEnvironment();
dotenv.config({ path: "./.env", quiet: true });
dotenv.config({ path: `./.env.${environment}`, quiet: true });

const projectName = process.env.PROJECT_NAME;
const awsAccountId = process.env.AWS_ACCOUNT_ID;
const awsRegion = process.env.AWS_REGION;

assert.ok(projectName, "PROJECT_NAME is required");
assert.ok(awsAccountId, "AWS_ACCOUNT_ID is required");
assert.ok(awsRegion, "AWS_REGION is required");

// The web, auth, data and api stacks come in the infra phase.
new cdk.App();
