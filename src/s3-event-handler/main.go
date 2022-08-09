// s3-object-event-handler
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func handler(ctx context.Context, s3Event events.S3Event) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := dynamodb.NewFromConfig(cfg)

	for _, record := range s3Event.Records {
		s3Object := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3Object.Bucket.Name, s3Object.Object.Key)

		// insert s3 object meta into a DynamoDB table
		client.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String(os.Getenv("TABLE_NAME")),
			Item: map[string]types.AttributeValue{
				"pk":        &types.AttributeValueMemberS{Value: s3Object.Bucket.Name},
				"sk":        &types.AttributeValueMemberS{Value: s3Object.Object.Key},
				"timestamp": &types.AttributeValueMemberS{Value: time.Now().String()},
			},
		})

	}
}

func main() {
	lambda.Start(handler)
}
