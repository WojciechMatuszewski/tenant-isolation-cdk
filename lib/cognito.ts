import * as cdk from "@aws-cdk/core";
import * as cognito from "@aws-cdk/aws-cognito";
import { GolangLambda } from "./lambda";

export class Cognito extends cdk.Construct {
  userPoolClientId: string;
  userPoolId: string;

  constructor(scope: cdk.Construct, id: string) {
    super(scope, id);

    const userPool = new cognito.UserPool(this, "userPool", {
      autoVerify: {
        email: true
      },
      signInAliases: {
        email: true
      },
      passwordPolicy: {
        minLength: 6,
        requireDigits: false,
        requireLowercase: false,
        requireSymbols: false,
        requireUppercase: false
      },
      selfSignUpEnabled: true,
      customAttributes: {
        tenant: new cognito.StringAttribute({ mutable: false })
      }
    });

    const userPoolClient = new cognito.UserPoolClient(this, "userPoolClient", {
      userPool,
      generateSecret: false,
      supportedIdentityProviders: [
        cognito.UserPoolClientIdentityProvider.COGNITO
      ],
      authFlows: {
        adminUserPassword: false,
        userPassword: true,
        userSrp: true
      }
    });

    const preSignUpTrigger = new GolangLambda(this, "preSignUpTrigger", {
      function: "verifier"
    });
    userPool.addTrigger(
      cognito.UserPoolOperation.PRE_SIGN_UP,
      preSignUpTrigger
    );

    this.userPoolClientId = userPoolClient.userPoolClientId;
    this.userPoolId = userPool.userPoolId;
  }
}
