package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQS client interface to allow mocking
type sqsClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// handle receiving each SQS message
func HandleMessage(ctx context.Context, handle string, body string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc := sqs.NewFromConfig(cfg)
	if err := processMessage(ctx, svc, handle, body); err != nil {
		log.Fatalf("unable to process message, %v", err)
	}
}

func processMessage(ctx context.Context, svc sqsClient, handle string, body string) error {
	var notifierQueue = os.Getenv("NOTIFIER_QUEUE")
	if notifierQueue == "" {
		log.Fatal("unable to get NOTIFIER_QUEUE")
	}
	var webmentionQueue = os.Getenv("WEBMENTION_QUEUE")
	if webmentionQueue == "" {
		log.Fatal("unable to get WEBMENTION_QUEUE")
	}
	if err := sendMessage(ctx, svc, notifierQueue, body); err != nil {
		return err
	}
	if err := sendMessage(ctx, svc, webmentionQueue, body); err != nil {
		return err
	}
	return deleteMessage(ctx, svc, handle)
}

func sendMessage(ctx context.Context, svc sqsClient, queue string, body string) error {
	_, err := svc.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queue,
		MessageBody: &body,
	})
	return err
}

// delete SQS message after it's handled
func deleteMessage(ctx context.Context, svc sqsClient, handle string) error {
	var incomingQueue = os.Getenv("INCOMING_QUEUE")
	if incomingQueue == "" {
		log.Fatal("unable to get INCOMING_QUEUE")
	}
	_, err := svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &incomingQueue,
		ReceiptHandle: &handle,
	})
	return err
}
