import * as cdk from "aws-cdk-lib/core";
import { Construct } from "constructs";
import { CfnOutput, Duration } from "aws-cdk-lib";
import type { ITable } from "aws-cdk-lib/aws-dynamodb";
import { type IRole, Policy, PolicyStatement } from "aws-cdk-lib/aws-iam";
import { GoFunction } from "./common/api/go-function";
import type { Environment } from "../utils/environment.util";

/**
 * Gemini calls a day across all users, each environment. Both share one
 * Google project, whose free tier allows 500, so together they stay under it
 * and Google never refuses for the day.
 */
export const GEMINI_DAILY_LIMIT: Record<Environment, number> = {
  development: 50,
  production: 400,
};

interface ApiStackProps extends cdk.StackProps {
  projectName: string;
  environment: Environment;
  table: ITable;
  /** The identity pool role the signed-in user's credentials carry. */
  authenticatedRole: IRole;
}

export class ApiStack extends cdk.Stack {
  readonly api: GoFunction;

  constructor(scope: Construct, id: string, props: ApiStackProps) {
    super(scope, id, props);

    const { projectName, environment, table, authenticatedRole } = props;

    // Picks wait on Gemini, then TMDB.
    this.api = new GoFunction(this, "Api", {
      projectName: projectName,
      environment: environment,
      command: "api",
      timeout: Duration.seconds(20),
      secrets: ["tmdb-token", "gemini-key"],
      environmentVariables: {
        TABLE_NAME: table.tableName,
        GEMINI_DAILY_LIMIT: String(GEMINI_DAILY_LIMIT[environment]),
      },
    });

    // Items and queries (a user's favourites), never a scan. A transaction
    // is checked against each of its items' actions; deleting an account
    // deletes its items in batches.
    table.grant(
      this.api.function,
      "dynamodb:GetItem",
      "dynamodb:Query",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:ConditionCheckItem",
      "dynamodb:BatchWriteItem",
    );

    // The browser invokes the function directly with identity-pool
    // credentials. A Policy owned by this stack, rather than
    // `grantInvoke(role)`, keeps the dependency one-way: grantInvoke would
    // write the statement into the auth stack's role and make the two stacks
    // reference each other.
    new Policy(this, "InvokeFunctions", {
      roles: [authenticatedRole],
      statements: [
        new PolicyStatement({
          actions: ["lambda:InvokeFunction"],
          resources: [this.api.function.functionArn],
        }),
      ],
    });

    new CfnOutput(this, "ApiFunctionName", {
      value: this.api.function.functionName,
    });
  }
}
