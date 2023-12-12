package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func lambdaHandler(ctx context.Context, sqsEvent events.SQSEvent) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	handler := MicroService{
		HTTPC: &http.Client{},
		SQSC:  sqs.NewFromConfig(cfg),
	}

	for _, message := range sqsEvent.Records {
		handler.HandleMessage(ctx, message.ReceiptHandle, message.Body)
	}
	return nil
}

func main() {
	lambda.Start(lambdaHandler)
}
