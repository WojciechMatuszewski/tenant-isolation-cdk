package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
)

func main() {
	lambda.Start(handler)
}

const (
	USER_POOL_ID        = "xxx"
	USER_POOL_CLIENT_ID = "xxx"
	REGION              = "xxx"
)

func handler(ctx context.Context,
	event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {

	keySetURL := fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v/.well-known/jwks.json", REGION, USER_POOL_ID)
	keyset, err := jwk.Fetch(ctx, keySetURL)
	if err != nil {
		fmt.Println("failed to fetch the keyset", err.Error())
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}

	token := event.AuthorizationToken
	parsedToken, err := jwt.Parse(
		[]byte(token),
		jwt.WithKeySet(keyset),
		jwt.WithValidate(true),
		jwt.WithIssuer(fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v", REGION, USER_POOL_ID)),
		jwt.WithClaimValue("client_id", USER_POOL_CLIENT_ID),
		jwt.WithClaimValue("token_use", "access"),
	)
	if err != nil {
		fmt.Println("failed to parse the token", err.Error())
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}

	username, found := parsedToken.Get("username")
	if !found {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Integrity issue, username not found")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}

	svc := cognitoidentityprovider.NewFromConfig(cfg)
	user, err := svc.GetUser(ctx, &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(token),
	})
	if err != nil {
		fmt.Println("failed to fetch the user")
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}

	var tentantID string
	for _, attribute := range user.UserAttributes {
		if *attribute.Name == "custom:tenant" {
			tentantID = *attribute.Value
			break
		}
	}
	if tentantID == "" {
		fmt.Println("tenant not within user attributes")
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}

	fmt.Println(event.MethodArn)
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: tentantID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Effect:   "Allow",
					Action:   []string{"execute-api:Invoke"},
					Resource: []string{event.MethodArn},
				},
			},
		},
		Context: map[string]interface{}{
			"username": username,
			"tenant":   tentantID,
		},
	}, nil
}
