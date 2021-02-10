import * as cdk from "@aws-cdk/core";
import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as apigwIntegrations from "@aws-cdk/aws-apigatewayv2-integrations";
import * as lambda from "@aws-cdk/aws-lambda";
import * as dynamo from "@aws-cdk/aws-dynamodb";
import { getFunctionPath } from "./utils/utils";
import { GolangLambda } from "./lambda";
import * as iam from "@aws-cdk/aws-iam";

export class API extends cdk.Construct {
  apiEndpoint: string;

  constructor(
    scope: cdk.Construct,
    id: string,
    props: { userPoolClientId: string; userPoolId: string }
  ) {
    super(scope, id);

    const dataTable = new dynamo.Table(this, "dataTable", {
      partitionKey: { name: "pk", type: dynamo.AttributeType.STRING }
    });

    const api = new apigw.HttpApi(this, "api", {
      corsPreflight: {
        allowOrigins: ["*"],
        allowMethods: [
          apigw.HttpMethod.GET,
          apigw.HttpMethod.POST,
          apigw.HttpMethod.OPTIONS
        ],
        allowHeaders: ["*"]
      }
    });
    this.apiEndpoint = api.apiEndpoint;

    const authorizerLambda = new GolangLambda(this, "authorizerLambda", {
      function: "authorizer"
    });
    authorizerLambda.grantInvoke(
      new iam.ServicePrincipal("apigateway.amazonaws.com")
    );

    const authorizer = new apigw.CfnAuthorizer(this, "authorizer", {
      apiId: api.httpApiId,
      authorizerType: "REQUEST",
      identitySource: ["$request.header.Authorization"],
      name: "JWT_AUTHORIZER",
      authorizerUri: `arn:aws:apigateway:${
        cdk.Stack.of(this).region
      }:lambda:path/2015-03-31/functions/${
        authorizerLambda.functionArn
      }/invocations`,
      authorizerPayloadFormatVersion: "1.0",
      authorizerResultTtlInSeconds: 0
    });

    const role = new iam.Role(this, "getterTenantRole", {
      assumedBy: new iam.AccountRootPrincipal(),
      inlinePolicies: {
        allowDynamo: new iam.PolicyDocument({
          statements: [
            new iam.PolicyStatement({
              effect: iam.Effect.ALLOW,
              actions: ["dynamodb:GetItem"],
              resources: [dataTable.tableArn]
            })
          ]
        })
      }
    });

    const getterFunction = new GolangLambda(this, "getterFunction", {
      function: "getter",
      env: {
        TABLE_ARN: dataTable.tableArn,
        TABLE_NAME: dataTable.tableName,
        ROLE_ARN: role.roleArn
      }
    });
    getterFunction.addToRolePolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: ["sts:AssumeRole"],
        resources: [role.roleArn]
      })
    );
    const getterRoute = new apigw.HttpRoute(this, "helloRoute", {
      integration: new apigwIntegrations.LambdaProxyIntegration({
        handler: getterFunction,
        payloadFormatVersion: apigw.PayloadFormatVersion.VERSION_2_0
      }),
      httpApi: api,
      routeKey: apigw.HttpRouteKey.with("/{tenantID}", apigw.HttpMethod.GET)
    });
    const getterRouteCfn = getterRoute.node.defaultChild as apigw.CfnRoute;
    getterRouteCfn.authorizationType = "CUSTOM";
    getterRouteCfn.authorizerId = authorizer.ref;

    const seederFunction = new GolangLambda(this, "seederFunction", {
      function: "seeder",
      env: {
        TABLE_NAME: dataTable.tableName
      }
    });
    dataTable.grantWriteData(seederFunction);
    const seederRoute = new apigw.HttpRoute(this, "seederRoute", {
      integration: new apigwIntegrations.LambdaProxyIntegration({
        handler: seederFunction,
        payloadFormatVersion: apigw.PayloadFormatVersion.VERSION_2_0
      }),
      httpApi: api,
      routeKey: apigw.HttpRouteKey.with("/seed", apigw.HttpMethod.GET)
    });
    const seederRouteCfn = seederRoute.node.defaultChild as apigw.CfnRoute;
    seederRouteCfn.authorizationType = "CUSTOM";
    seederRouteCfn.authorizerId = authorizer.ref;
  }
}
