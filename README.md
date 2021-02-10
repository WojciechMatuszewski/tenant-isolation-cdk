# About

This repo contains a very rough and messy implementation of the pattern from [this article](https://aws.amazon.com/blogs/apn/isolating-saas-tenants-with-dynamically-generated-iam-policies/)

I did not focus at the clarity of the code at all.

## Learnings

### ID Token vs Access Token

- do not send ID Token through the wire to your services. This [auth0 article](https://auth0.com/docs/tokens) mentions that
  > ID tokens are JSON web tokens meant for use by the application only. [...] Do **not** use ID tokens to gain access to an API.

#### Adding custom scopes to the access token

- since you should not send the ID token, you will be sending AccessToken
- Access Token does not have the scopes by default, but you still want to get them

- one way would be to make a call to cognito

- another would be to parse Cognito Access Token using custom authorizer

- to create a custom authorizer, you will need the combination of jwts and jwks

- the `jwt/go` seems to abandoned, use the `jwx` library, at least for jwks

#### Authorizer response format

- there are 2 payload versions you can pick from for your authorizer (similarly to the APIGW request / response format)
- you the **response from the authorizer has to be in a specific format and not missing any fields**. Otherwise a 500 error will be thrown. Refer to the docs

### IAM Sessions policies

- session policies are the policies you pass when you want to assume a role

- remember that the overall permissions that you will get, is an intersection of the permissions of the role that you want to assume AND the policy you specify

- in the case of lambda, the execution role has nothing to do with the session policy

### DynamoDB granular permissions

- you can create granular IAM permissions for dynamodb operations

- we are using `ForAllValues:StringEquals` due to the fact that `dynamodb:LeadingKeys` takes multiple keys
