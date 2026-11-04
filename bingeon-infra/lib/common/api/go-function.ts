import { execFileSync } from "node:child_process";
import * as path from "node:path";
import { DockerImage, Duration, RemovalPolicy, Stack } from "aws-cdk-lib";
import { PolicyStatement } from "aws-cdk-lib/aws-iam";
import { Architecture, Code, Function as LambdaFunction, Runtime } from "aws-cdk-lib/aws-lambda";
import { LogGroup, RetentionDays } from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";
import type { Environment } from "../../../utils/environment.util";
import { type Secret, secretParameter, secretVariable } from "../../../utils/secrets.util";

const LAMBDA_ROOT = path.join(__dirname, "../../../../bingeon-lambda");

export interface GoFunctionProps {
  projectName: string;
  environment: Environment;
  /** The command under bingeon-lambda/cmd to build, e.g. "api". */
  command: string;
  environmentVariables?: Record<string, string>;
  /**
   * The keys it reads: each parameter's name is passed as <SECRET>_PARAM
   * (see secretVariable), and it may read that parameter and no other.
   */
  secrets?: ReadonlyArray<Secret>;
  timeout: Duration;
  /** MB, which also sets its share of a CPU. */
  memorySize?: number;
}

/**
 * A Lambda built from bingeon-lambda/cmd/<command> at synth time.
 *
 * Builds with the local Go toolchain (cross-compiled for arm64 Linux) so
 * synth needs no Docker; the Docker image is only a fallback CDK reaches for
 * if local bundling reports it could not run. No reserved concurrency: the
 * account's limit is 10, and AWS keeps 10 unreserved.
 */
export class GoFunction extends Construct {
  readonly function: LambdaFunction;

  constructor(scope: Construct, id: string, props: GoFunctionProps) {
    super(scope, id);

    const { projectName, environment, command, environmentVariables, secrets = [], timeout, memorySize = 256 } = props;

    const functionName = `${projectName}-${environment}-${command}`;

    const logGroup = new LogGroup(this, "Logs", {
      logGroupName: `/aws/lambda/${functionName}`,
      retention: RetentionDays.ONE_MONTH,
      removalPolicy: RemovalPolicy.DESTROY,
    });

    const parameters = secrets.map((secret) => ({
      secret,
      name: secretParameter(projectName, environment, secret),
    }));

    this.function = new LambdaFunction(this, "Function", {
      functionName: functionName,
      runtime: Runtime.PROVIDED_AL2023,
      architecture: Architecture.ARM_64,
      handler: "bootstrap",
      memorySize: memorySize,
      timeout: timeout,
      logGroup: logGroup,
      environment: {
        ...environmentVariables,
        ...Object.fromEntries(parameters.map(({ secret, name }) => [secretVariable(secret), name])),
      },
      code: Code.fromAsset(LAMBDA_ROOT, {
        bundling: {
          image: DockerImage.fromRegistry("golang:1.26"),
          environment: { GOOS: "linux", GOARCH: "arm64", CGO_ENABLED: "0" },
          command: ["bash", "-c", `go build -trimpath -ldflags="-s -w" -o /asset-output/bootstrap ./cmd/${command}`],
          local: {
            tryBundle(outputDir: string) {
              execFileSync(
                "go",
                ["build", "-trimpath", "-ldflags=-s -w", "-o", path.join(outputDir, "bootstrap"), `./cmd/${command}`],
                {
                  cwd: LAMBDA_ROOT,
                  stdio: "inherit",
                  env: { ...process.env, GOOS: "linux", GOARCH: "arm64", CGO_ENABLED: "0" },
                },
              );
              return true;
            },
          },
        },
      }),
    });

    // The default aws/ssm key's policy already lets SSM decrypt for the
    // account, so no KMS grant is needed.
    if (parameters.length > 0) {
      const stack = Stack.of(this);
      this.function.addToRolePolicy(
        new PolicyStatement({
          actions: ["ssm:GetParameter"],
          resources: parameters.map(({ name }) =>
            stack.formatArn({ service: "ssm", resource: "parameter", resourceName: name.slice(1) }),
          ),
        }),
      );
    }
  }
}
