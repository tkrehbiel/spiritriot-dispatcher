package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaHandler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		HandleMessage(ctx, message.ReceiptHandle, message.Body)
	}
	return nil
}

func main() {
	lambda.Start(lambdaHandler)
}
