import * as lambda from "@aws-cdk/aws-lambda";
import * as cdk from "@aws-cdk/core";
import { getFunctionPath } from "./utils/utils";

export class GolangLambda extends lambda.Function {
  constructor(
    scope: cdk.Construct,
    id: string,
    props: { function: string; env?: Record<string, string> }
  ) {
    super(scope, id, {
      code: lambda.Code.fromAsset(getFunctionPath(props.function)),
      handler: "main",
      runtime: lambda.Runtime.GO_1_X,
      environment: props.env
    });
  }
}
