import { Duration, RemovalPolicy } from "aws-cdk-lib";
import {
  AllowedMethods,
  CachePolicy,
  CacheCookieBehavior,
  CacheHeaderBehavior,
  CacheQueryStringBehavior,
  Distribution,
  Function as CloudFrontFunction,
  FunctionCode,
  FunctionEventType,
  FunctionRuntime,
  HttpVersion,
  PriceClass,
  ViewerProtocolPolicy,
} from "aws-cdk-lib/aws-cloudfront";
import { FunctionUrlOrigin, S3BucketOrigin } from "aws-cdk-lib/aws-cloudfront-origins";
import type { IFunctionUrl } from "aws-cdk-lib/aws-lambda";
import { BlockPublicAccess, Bucket, BucketEncryption, ObjectOwnership } from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import type { Environment } from "../../../utils/environment.util";

/**
 * The query parameters bingeon-lambda's cmd/catalog reads, the only ones in
 * the cache key and the only ones it receives: anything else a caller adds
 * can't make a new cache entry. A parameter the catalogue starts reading
 * must be added here.
 */
export const CATALOG_QUERY = ["id", "list", "genre", "q", "page"];

export interface SiteProps {
  projectName: string;
  environment: Environment;
  /** The catalog function's URL (AWS_IAM), served at /catalog/*. */
  catalog: IFunctionUrl;
}

/**
 * The SPA in a private bucket and the catalogue, behind one CloudFront.
 *
 * Kept cheap: the bucket isn't versioned, the cheapest price class, no
 * logging. Hashed assets are cached for a year. The catalogue is cached for
 * as long as each reply's s-maxage says, and not at all without one.
 */
export class Site extends Construct {
  readonly bucket: Bucket;
  readonly distribution: Distribution;

  constructor(scope: Construct, id: string, props: SiteProps) {
    super(scope, id);

    const { projectName, environment, catalog } = props;

    const retained = environment === "production";

    this.bucket = new Bucket(this, "SiteBucket", {
      encryption: BucketEncryption.S3_MANAGED,
      blockPublicAccess: BlockPublicAccess.BLOCK_ALL,
      objectOwnership: ObjectOwnership.BUCKET_OWNER_ENFORCED,
      enforceSSL: true,
      removalPolicy: retained ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY,
      autoDeleteObjects: !retained,
      // An upload a deploy didn't finish would otherwise be billed for good.
      lifecycleRules: [{ id: "abort-incomplete-uploads", abortIncompleteMultipartUploadAfter: Duration.days(1) }],
    });

    // SPA fallback, as a viewer-request rewrite of extensionless paths only,
    // rather than mapping 403/404 to index.html for the whole distribution:
    // that would also answer a missing asset (a stale lazy chunk) with
    // index.html and a 200, and the catalogue's 404s with the app.
    const spaRouting = new CloudFrontFunction(this, "SpaRouting", {
      comment: `${projectName}-${environment}-web spa routing`,
      runtime: FunctionRuntime.JS_2_0,
      code: FunctionCode.fromInline(`
function handler(event) {
  var request = event.request;
  var lastSegment = request.uri.substring(request.uri.lastIndexOf('/') + 1);
  if (lastSegment.indexOf('.') === -1) {
    request.uri = '/index.html';
  }
  return request;
}
      `),
    });

    const catalogCache = new CachePolicy(this, "CatalogCache", {
      cachePolicyName: `${projectName}-${environment}-catalog`,
      comment: "Keyed only on the parameters the catalogue reads, for as long as it says",
      minTtl: Duration.seconds(0),
      defaultTtl: Duration.seconds(0),
      maxTtl: Duration.days(1),
      queryStringBehavior: CacheQueryStringBehavior.allowList(...CATALOG_QUERY),
      headerBehavior: CacheHeaderBehavior.none(),
      cookieBehavior: CacheCookieBehavior.none(),
      enableAcceptEncodingGzip: true,
      enableAcceptEncodingBrotli: true,
    });

    this.distribution = new Distribution(this, "Distribution", {
      comment: `${projectName}-${environment}-web`,
      defaultRootObject: "index.html",
      priceClass: PriceClass.PRICE_CLASS_100,
      httpVersion: HttpVersion.HTTP2_AND_3,
      defaultBehavior: {
        origin: S3BucketOrigin.withOriginAccessControl(this.bucket),
        viewerProtocolPolicy: ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        functionAssociations: [{ function: spaRouting, eventType: FunctionEventType.VIEWER_REQUEST }],
      },
      additionalBehaviors: {
        // Signed by CloudFront (origin access control), so the function URL
        // answers nothing else. No origin request policy: the function gets
        // only what is in the cache key.
        "/catalog/*": {
          origin: FunctionUrlOrigin.withOriginAccessControl(catalog),
          viewerProtocolPolicy: ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
          allowedMethods: AllowedMethods.ALLOW_GET_HEAD,
          cachePolicy: catalogCache,
        },
      },
    });
  }
}
