import * as cdk from "@aws-cdk/core";
import { API } from "./api";
import { Cognito } from "./cognito";

export class Stack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const { userPoolClientId, userPoolId } = new Cognito(this, "cognito");
    const { apiEndpoint } = new API(this, "api", {
      userPoolClientId,
      userPoolId
    });

    new cdk.CfnOutput(this, "apiEndpoint", {
      value: apiEndpoint
    });

    new cdk.CfnOutput(this, "userPoolClientId", {
      value: userPoolClientId
    });

    new cdk.CfnOutput(this, "userPoolId", {
      value: userPoolId
    });

    new cdk.CfnOutput(this, "region", {
      value: this.region
    });
  }
}
