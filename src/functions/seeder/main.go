package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	lambda.Start(handler)
}

type Item struct {
	PK   string `dynamodbav:"pk"`
	Data string `dynamodbav:"data"`
}

func handler(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("loading config failed", err.Error())
		panic(err.Error())
	}

	svc := dynamodb.NewFromConfig(cfg)
	item1 := Item{
		PK:   "1",
		Data: "tenant1",
	}
	av1, err := attributevalue.MarshalMap(item1)
	if err != nil {
		fmt.Println("av1 failed")
		panic(err.Error())
	}

	item2 := Item{
		PK:   "2",
		Data: "tenant2",
	}
	av2, err := attributevalue.MarshalMap(item2)
	if err != nil {
		fmt.Println("av2 failed")
		panic(err.Error())
	}

	table := os.Getenv("TABLE_NAME")
	_, err = svc.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			table: {
				{
					PutRequest: &types.PutRequest{
						Item: av1,
					},
				},
				{
					PutRequest: &types.PutRequest{
						Item: av2,
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println("read to dynamo failed", err.Error())
		panic(err.Error())
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       "Works!",
	}, nil

}
