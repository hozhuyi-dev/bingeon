import * as cdk from "aws-cdk-lib/core";
import { Match, Template } from "aws-cdk-lib/assertions";
import { describe, expect, test } from "vitest";
import { CiStack } from "../lib/ci-stack";

const env = { account: "123456789012", region: "ap-southeast-1" };

function synth() {
  const app = new cdk.App();
  const stack = new CiStack(app, "ci", {
    env,
    projectName: "bingeon",
    repository: "owner@1/bingeon@2",
    environments: ["production"],
  });
  return Template.fromStack(stack);
}

type Statement = { Action: string | Array<string>; Resource: unknown };
type Policy = { Properties: { Roles: Array<{ Ref: string }>; PolicyDocument: { Statement: Array<Statement> } } };

/** The statements of the role named `roleName`. */
function statementsOf(template: Template, roleName: string) {
  const roles = template.findResources("AWS::IAM::Role", { Properties: { RoleName: roleName } });
  const [roleId] = Object.keys(roles);
  expect(roleId).toBeDefined();
  return Object.values(template.findResources("AWS::IAM::Policy") as Record<string, Policy>)
    .filter((policy) => policy.Properties.Roles.some((role) => role.Ref === roleId))
    .flatMap((policy) => policy.Properties.PolicyDocument.Statement);
}

describe("ci", () => {
  const template = synth();

  test("uses the account's GitHub provider rather than making one", () => {
    template.resourceCountIs("AWS::IAM::OIDCProvider", 0);
    template.resourceCountIs("Custom::AWSCDKOpenIdConnectProvider", 0);
  });

  test("has only production's role, which only its GitHub environment can assume", () => {
    template.resourceCountIs("AWS::IAM::Role", 1);
    template.hasResourceProperties("AWS::IAM::Role", {
      RoleName: "bingeon-production-github",
      AssumeRolePolicyDocument: {
        Statement: [
          Match.objectLike({
            Action: "sts:AssumeRoleWithWebIdentity",
            Principal: {
              Federated: {
                "Fn::Join": ["", Match.arrayWith([":oidc-provider/token.actions.githubusercontent.com"])],
              },
            },
            Condition: {
              StringEquals: {
                "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
                "token.actions.githubusercontent.com:sub": "repo:owner@1/bingeon@2:environment:production",
              },
            },
          }),
        ],
      },
    });
  });

  test("scopes its own resources to production", () => {
    const statements = JSON.stringify(statementsOf(template, "bingeon-production-github"));
    expect(statements).toContain("stack/bingeon-production-*/*");
    expect(statements).toContain("bingeon-production-web-*");
    expect(statements).not.toContain("development");
  });

  test("grants nothing account-wide but reading stacks and invalidating caches", () => {
    const wide = statementsOf(template, "bingeon-production-github").filter((s) =>
      JSON.stringify(s.Resource).match(/"\*"|stack\/\*\/\*|distribution\/\*/),
    );
    expect(wide.flatMap((s) => s.Action).sort()).toEqual([
      "cloudformation:DescribeStacks",
      "cloudformation:GetTemplate",
      "cloudformation:GetTemplateSummary",
      "cloudformation:ListStacks",
      "cloudfront:CreateInvalidation",
    ]);
  });
});
