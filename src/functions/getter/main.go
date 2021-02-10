package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type User struct {
	TenantID string `json:"tenant"`
	Username string `json:"username`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	user := getUserFromContext(event.RequestContext.Authorizer.Lambda)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return respond(http.StatusInternalServerError, fmt.Sprint("failed to load default config", err.Error())), nil
	}

	tenantFromPath, found := event.PathParameters["tenantID"]
	if !found {
		return respond(http.StatusBadRequest, fmt.Sprint("tenantID not found in path parameters")), nil
	}

	svc := sts.NewFromConfig(cfg)

	policy := policyForTentant(getEnv("TABLE_ARN"), user.TenantID)
	fmt.Println(policy)
	out, err := svc.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(getEnv("ROLE_ARN")),
		RoleSessionName: aws.String(fmt.Sprintf("tenant%v", user.TenantID)),
		Policy:          aws.String(policy),
	})
	if err != nil {
		return respond(http.StatusInternalServerError, fmt.Sprint("failed to assume the role", err.Error())), nil
	}

	newConfig, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(
		credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     *out.Credentials.AccessKeyId,
				SecretAccessKey: *out.Credentials.SecretAccessKey,
				SessionToken:    *out.Credentials.SessionToken,
				Source:          "Source",
			},
		},
	))
	if err != nil {
		return respond(http.StatusInternalServerError, fmt.Sprint("failed to load the credentials", err.Error())), nil
	}

	client := dynamodb.NewFromConfig(newConfig)
	getItemOut, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(getEnv("TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: tenantFromPath},
		},
	})
	if err != nil {
		return respond(http.StatusInternalServerError, fmt.Sprint("failed to fetch dynamo item", err.Error())), nil
	}

	type Item struct {
		PK   string `dynamoav:"pk"`
		Data string `dynamoab:"data"`
	}
	var item Item
	err = attributevalue.UnmarshalMap(getItemOut.Item, &item)
	if err != nil {
		return respond(http.StatusInternalServerError, fmt.Sprint("failed to unmarshal output from dynamo", err.Error())), nil
	}

	itemB, err := json.Marshal(item)
	if err != nil {
		return respond(http.StatusInternalServerError, fmt.Sprint("failed to unmarshal dynamodb item", err.Error())), nil
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       string(itemB),
	}, nil
}

func getUserFromContext(authContext map[string]interface{}) User {
	user := User{}

	tenantID, found := authContext["tenant"].(string)
	if !found {
		panic("tenantID not found")
	}
	user.TenantID = tenantID

	username, found := authContext["username"].(string)
	if !found {
		panic("Username not found")
	}
	user.Username = username

	return user
}

func getEnv(key string) string {
	value, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Sprintf("key %v not found", key))
	}

	return value
}

func respond(statusCode int, body string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Body:       body,
	}
}

func policyForTentant(tableArn, tenantID string) string {
	return fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Action": ["dynamodb:GetItem"],
					"Resource": "%v",
					"Effect": "Allow",
					"Condition": {
						"ForAllValues:StringEquals": {
							"dynamodb:LeadingKeys": [
								"%v"
							]
						}
					}
				}
			]
		}`, tableArn, tenantID)
}
