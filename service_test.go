package main

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcess(t *testing.T) {
	os.Setenv("INCOMING_QUEUE", "queue1")
	os.Setenv("NOTIFIER_QUEUE", "queue2")
	os.Setenv("WEBMENTION_QUEUE", "queue3")

	handle := "eventHandle"
	body := "{ \"url\": \"anything\" }"

	ctx := context.Background()

	mockClient := new(MockSQSClient)
	mockClient.On("SendMessage", ctx, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
		return *input.QueueUrl == "queue2" && *input.MessageBody == body
	}), mock.Anything).Return(nil, nil).Once()
	mockClient.On("SendMessage", ctx, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
		return *input.QueueUrl == "queue3" && *input.MessageBody == body
	}), mock.Anything).Return(nil, nil).Once()
	mockClient.On("DeleteMessage", ctx, mock.MatchedBy(func(input *sqs.DeleteMessageInput) bool {
		return *input.QueueUrl == "queue1" && *input.ReceiptHandle == handle
	}), mock.Anything).Return(nil, nil).Once()

	err := processMessage(ctx, mockClient, handle, body)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}
