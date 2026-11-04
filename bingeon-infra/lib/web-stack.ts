import * as cdk from "aws-cdk-lib/core";
import { Construct } from "constructs";
import { CfnOutput, Duration } from "aws-cdk-lib";
import { FunctionUrlAuthType } from "aws-cdk-lib/aws-lambda";
import { GoFunction } from "./common/api/go-function";
import { Site } from "./common/web/site";
import type { Environment } from "../utils/environment.util";

interface WebStackProps extends cdk.StackProps {
  projectName: string;
  environment: Environment;
}

/**
 * What anyone reaches: the site and the catalogue behind it. The catalog
 * function lives here, not in the api stack, because the distribution needs
 * it and sign-in needs the distribution's URL: anywhere else the stacks
 * would depend on each other.
 */
export class WebStack extends cdk.Stack {
  readonly site: Site;
  readonly catalog: GoFunction;
  /** "https://d123.cloudfront.net", which sign-in takes. */
  readonly siteOrigin: string;

  constructor(scope: Construct, id: string, props: WebStackProps) {
    super(scope, id, props);

    const { projectName, environment } = props;

    // Anyone can call it, through CloudFront, and every uncached call is
    // billed, so it reads the TMDB token and nothing else, and a stuck call
    // is cut short.
    this.catalog = new GoFunction(this, "Catalog", {
      projectName: projectName,
      environment: environment,
      command: "catalog",
      timeout: Duration.seconds(5),
      secrets: ["tmdb-token"],
    });
    const catalogUrl = this.catalog.function.addFunctionUrl({ authType: FunctionUrlAuthType.AWS_IAM });

    this.site = new Site(this, "Site", {
      projectName: projectName,
      environment: environment,
      catalog: catalogUrl,
    });

    this.siteOrigin = `https://${this.site.distribution.distributionDomainName}`;

    new CfnOutput(this, "SiteUrl", {
      value: this.siteOrigin,
    });

    new CfnOutput(this, "DistributionId", {
      value: this.site.distribution.distributionId,
    });

    new CfnOutput(this, "SiteBucketName", {
      value: this.site.bucket.bucketName,
    });
  }
}
