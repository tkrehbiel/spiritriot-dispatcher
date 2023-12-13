package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func lambdaHandlerSNS(ctx context.Context, snsEvent events.SNSEvent) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	handler := MicroService{
		HTTPClient: &http.Client{},
		SNSClient:  sns.NewFromConfig(cfg),
	}

	for _, message := range snsEvent.Records {
		handler.HandleMessage(ctx, message.SNS.Message)
	}
	return nil
}

func main() {
	lambda.Start(lambdaHandlerSNS)
}
