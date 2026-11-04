import * as cdk from "aws-cdk-lib/core";
import { Match, Template } from "aws-cdk-lib/assertions";
import { describe, expect, test } from "vitest";
import { ApiStack, GEMINI_DAILY_LIMIT } from "../lib/api-stack";
import { AuthStack } from "../lib/auth-stack";
import { DataStack } from "../lib/data-stack";
import { WebStack } from "../lib/web-stack";
import { CATALOG_QUERY } from "../lib/common/web/site";
import { appUrls } from "../utils/app-urls.util";
import type { Environment } from "../utils/environment.util";

const env = { account: "123456789012", region: "ap-southeast-1" };

function synth(environment: Environment) {
  // Templates are all these tests read, so skip compiling the Go assets.
  const app = new cdk.App({ context: { "aws:cdk:bundling-stacks": [] } });

  // As bin/bingeon-infra.ts wires them: the site first, then localhost.
  const web = new WebStack(app, "web", { env, projectName: "bingeon", environment });
  const urls = appUrls(web.siteOrigin, {
    callbackUrls: ["http://localhost:4200/"],
    logoutUrls: ["http://localhost:4200/"],
  });

  const auth = new AuthStack(app, "auth", {
    env,
    projectName: "bingeon",
    environment,
    callbackUrls: urls.callbackUrls,
    logoutUrls: urls.logoutUrls,
    googleClientId: "client-id",
    googleClientSecret: "client-secret",
  });

  const data = new DataStack(app, "data", { env, projectName: "bingeon", environment });

  const api = new ApiStack(app, "api", {
    env,
    projectName: "bingeon",
    environment,
    table: data.data.table,
    authenticatedRole: auth.auth.authenticatedRole,
  });

  return {
    auth: Template.fromStack(auth),
    data: Template.fromStack(data),
    api: Template.fromStack(api),
    web: Template.fromStack(web),
  };
}

type Statement = { Action: string | Array<string>; Resource: unknown };
type Policy = { Properties: { Roles: Array<{ Ref: string }>; PolicyDocument: { Statement: Array<Statement> } } };

/** The policy statements attached to the role whose logical id starts with `rolePrefix`. */
function statementsOf(template: Template, rolePrefix: string) {
  const roleId = Object.keys(template.findResources("AWS::IAM::Role")).find((id) => id.startsWith(rolePrefix));
  expect(roleId).toBeDefined();
  return Object.values(template.findResources("AWS::IAM::Policy") as Record<string, Policy>)
    .filter((policy) => policy.Properties.Roles.some((role) => role.Ref === roleId))
    .flatMap((policy) => policy.Properties.PolicyDocument.Statement);
}

/** The web stack's CloudFront URL as another stack imports it, then `path`. */
function siteUrl(path: string) {
  return { "Fn::Join": ["", ["https://", { "Fn::ImportValue": Match.stringLikeRegexp("^web:") }, path]] };
}

const { auth, data, api, web } = synth("development");

describe("auth", () => {
  test("signs in and out at the site first, then localhost", () => {
    auth.hasResourceProperties("AWS::Cognito::UserPoolClient", {
      CallbackURLs: [siteUrl("/"), "http://localhost:4200/"],
      LogoutURLs: [siteUrl("/"), "http://localhost:4200/"],
      DefaultRedirectURI: siteUrl("/"),
    });
  });

  test("offers Google as the only provider, with no self sign-up", () => {
    auth.hasResourceProperties("AWS::Cognito::UserPoolClient", { SupportedIdentityProviders: ["Google"] });
    auth.hasResourceProperties("AWS::Cognito::UserPool", { AdminCreateUserConfig: { AllowAdminCreateUserOnly: true } });
  });

  test("asks Google only for the email", () => {
    auth.hasResourceProperties("AWS::Cognito::UserPoolIdentityProvider", {
      ProviderDetails: Match.objectLike({ authorize_scopes: "openid email" }),
      AttributeMapping: { email: "email", email_verified: "email_verified" },
    });
  });

  test("gives every signed-in user the one role, and guests nothing", () => {
    auth.hasResourceProperties("AWS::Cognito::IdentityPool", { AllowUnauthenticatedIdentities: false });
    auth.hasResourceProperties("AWS::Cognito::IdentityPoolRoleAttachment", { RoleMappings: Match.absent() });
    // The construct always makes a guest role; it can do nothing.
    expect(statementsOf(auth, "AuthIdentityPoolUnauthenticatedRole")).toEqual([]);
    const roles = auth.findResources("AWS::IAM::Role") as Record<string, { Properties: Record<string, unknown> }>;
    const guest = Object.entries(roles).find(([id]) => id.startsWith("AuthIdentityPoolUnauthenticatedRole"))?.[1];
    expect(guest?.Properties.ManagedPolicyArns).toBeUndefined();
    expect(guest?.Properties.Policies).toBeUndefined();
  });
});

describe("data", () => {
  test("stays on free-tier provisioned capacity, expiring items by `expires`", () => {
    data.hasResourceProperties("AWS::DynamoDB::Table", {
      BillingMode: Match.absent(),
      ProvisionedThroughput: { ReadCapacityUnits: 5, WriteCapacityUnits: 5 },
      TimeToLiveSpecification: { AttributeName: "expires", Enabled: true },
      GlobalSecondaryIndexes: Match.absent(),
      LocalSecondaryIndexes: Match.absent(),
    });
  });

  test("retains only the production table", () => {
    data.hasResource("AWS::DynamoDB::Table", { DeletionPolicy: "Delete" });
    synth("production").data.hasResource("AWS::DynamoDB::Table", {
      DeletionPolicy: "Retain",
      Properties: Match.objectLike({ DeletionProtectionEnabled: true }),
    });
  });
});

describe("api", () => {
  test("lets the signed-in role invoke only the api function", () => {
    const [functionId] = Object.keys(
      api.findResources("AWS::Lambda::Function", { Properties: { FunctionName: "bingeon-development-api" } }),
    );
    // The role's only policy, owned by the api stack.
    auth.resourceCountIs("AWS::IAM::Policy", 0);
    const policies = Object.values(
      api.findResources("AWS::IAM::Policy", { Properties: { PolicyName: Match.stringLikeRegexp("InvokeFunctions") } }),
    ) as Array<Policy>;
    expect(policies.flatMap((p) => p.Properties.PolicyDocument.Statement)).toEqual([
      { Action: "lambda:InvokeFunction", Effect: "Allow", Resource: { "Fn::GetAtt": [functionId, "Arn"] } },
    ]);
  });

  test("reads, queries and writes items but never scans, and reads its own two keys", () => {
    const statements = statementsOf(api, "Api");
    expect(statements.map((s) => s.Action)).toEqual([
      "ssm:GetParameter",
      [
        "dynamodb:GetItem",
        "dynamodb:Query",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem",
        "dynamodb:ConditionCheckItem",
        "dynamodb:BatchWriteItem",
      ],
    ]);
    const ssm = JSON.stringify(statements[0]?.Resource);
    expect(ssm).toContain("parameter/bingeon/development/tmdb-token");
    expect(ssm).toContain("parameter/bingeon/development/gemini-key");
  });

  test("is told its table and where its keys are", () => {
    api.hasResourceProperties("AWS::Lambda::Function", {
      FunctionName: "bingeon-development-api",
      Timeout: 20,
      MemorySize: 256,
      Environment: {
        Variables: {
          TABLE_NAME: Match.anyValue(),
          GEMINI_DAILY_LIMIT: "50",
          TMDB_TOKEN_PARAM: "/bingeon/development/tmdb-token",
          GEMINI_KEY_PARAM: "/bingeon/development/gemini-key",
        },
      },
    });
  });

  test("keeps both environments' Gemini calls under the shared free 500 a day", () => {
    expect(GEMINI_DAILY_LIMIT.development + GEMINI_DAILY_LIMIT.production).toBeLessThan(500);
    synth("production").api.hasResourceProperties("AWS::Lambda::Function", {
      FunctionName: "bingeon-production-api",
      Environment: { Variables: Match.objectLike({ GEMINI_DAILY_LIMIT: "400" }) },
    });
  });

  test("reserves no concurrency", () => {
    for (const template of [api, web]) {
      expect(JSON.stringify(template.toJSON())).not.toContain("ReservedConcurrentExecutions");
    }
  });
});

describe("catalog", () => {
  test("reads the TMDB token and nothing else", () => {
    const statements = statementsOf(web, "Catalog");
    expect(statements.map((s) => s.Action)).toEqual(["ssm:GetParameter"]);
    expect(JSON.stringify(statements[0]?.Resource)).toContain("parameter/bingeon/development/tmdb-token");
    web.hasResourceProperties("AWS::Lambda::Function", {
      FunctionName: "bingeon-development-catalog",
      Timeout: 5,
      Environment: { Variables: { TMDB_TOKEN_PARAM: "/bingeon/development/tmdb-token" } },
    });
  });

  test("answers only CloudFront's signed requests", () => {
    web.hasResourceProperties("AWS::Lambda::Url", { AuthType: "AWS_IAM" });
    web.hasResourceProperties("AWS::CloudFront::OriginAccessControl", {
      OriginAccessControlConfig: Match.objectLike({ OriginAccessControlOriginType: "lambda", SigningBehavior: "always" }),
    });
    for (const action of ["lambda:InvokeFunctionUrl", "lambda:InvokeFunction"]) {
      web.hasResourceProperties("AWS::Lambda::Permission", {
        Action: action,
        Principal: "cloudfront.amazonaws.com",
        SourceArn: Match.anyValue(),
      });
    }
  });

  test("is served at /catalog/*, GET only, cached by its own word and keyed on what it reads", () => {
    web.hasResourceProperties("AWS::CloudFront::Distribution", {
      DistributionConfig: Match.objectLike({
        CacheBehaviors: [
          Match.objectLike({
            PathPattern: "/catalog/*",
            AllowedMethods: ["GET", "HEAD"],
            ViewerProtocolPolicy: "redirect-to-https",
            OriginRequestPolicyId: Match.absent(),
            FunctionAssociations: Match.absent(),
          }),
        ],
      }),
    });
    web.hasResourceProperties("AWS::CloudFront::CachePolicy", {
      CachePolicyConfig: Match.objectLike({
        Name: "bingeon-development-catalog",
        DefaultTTL: 0,
        MinTTL: 0,
        MaxTTL: 86400,
        ParametersInCacheKeyAndForwardedToOrigin: Match.objectLike({
          QueryStringsConfig: { QueryStringBehavior: "whitelist", QueryStrings: CATALOG_QUERY },
          HeadersConfig: { HeaderBehavior: "none" },
          CookiesConfig: { CookieBehavior: "none" },
        }),
      }),
    });
  });
});

describe("web", () => {
  test("keeps no old builds: the bucket isn't versioned, and abandoned uploads go", () => {
    const buckets = Object.values(web.findResources("AWS::S3::Bucket")) as Array<{ Properties: Record<string, unknown> }>;
    expect(buckets).toHaveLength(1);
    expect(buckets[0]?.Properties.VersioningConfiguration).toBeUndefined();
    web.hasResourceProperties("AWS::S3::Bucket", {
      LifecycleConfiguration: {
        Rules: [Match.objectLike({ Id: "abort-incomplete-uploads", AbortIncompleteMultipartUpload: { DaysAfterInitiation: 1 } })],
      },
    });
  });

  test("serves at the cheapest price class, without logs", () => {
    web.hasResourceProperties("AWS::CloudFront::Distribution", {
      DistributionConfig: Match.objectLike({ PriceClass: "PriceClass_100", Logging: Match.absent() }),
    });
  });
});
