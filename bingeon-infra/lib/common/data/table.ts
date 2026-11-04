import { RemovalPolicy } from "aws-cdk-lib";
import { AttributeType, BillingMode, Table } from "aws-cdk-lib/aws-dynamodb";
import { Construct } from "constructs";
import type { Environment } from "../../../utils/environment.util";

/**
 * Read and write capacity units, each environment. DynamoDB's free tier is
 * 25 of each, provisioned, shared by every table in the account; bizzy's
 * three tables take 15, so bingeon's two take the other 10. Provisioned
 * capacity is also the cost cap: beyond it, requests are throttled. Items
 * are kept small for it (a heart is about 4 write units; see bingeon-lambda's
 * platform/table).
 */
export const CAPACITY_UNITS = 5;

export interface DataTableProps {
  projectName: string;
  environment: Environment;
}

/**
 * The single bingeon table: PK and SK, no index (see bingeon-lambda's
 * platform/table). Items that expire say when in `expires`.
 */
export class DataTable extends Construct {
  readonly table: Table;

  constructor(scope: Construct, id: string, props: DataTableProps) {
    super(scope, id);

    const { projectName, environment } = props;

    const retained = environment === "production";

    // No point-in-time recovery: favourites can be made again.
    this.table = new Table(this, "Table", {
      tableName: `${projectName}-${environment}`,
      partitionKey: { name: "PK", type: AttributeType.STRING },
      sortKey: { name: "SK", type: AttributeType.STRING },
      billingMode: BillingMode.PROVISIONED,
      readCapacity: CAPACITY_UNITS,
      writeCapacity: CAPACITY_UNITS,
      timeToLiveAttribute: "expires",
      deletionProtection: retained,
      removalPolicy: retained ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY,
    });
  }
}
