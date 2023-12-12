package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func matches(source string, target string, js string) bool {
	var mention Mention
	if err := json.Unmarshal([]byte(js), &mention); err != nil {
		return false
	}
	return source == mention.Source && target == mention.Target
}

func TestProcess(t *testing.T) {
	os.Setenv("INCOMING_QUEUE", "queue1")
	os.Setenv("NOTIFIER_QUEUE", "queue2")
	os.Setenv("WEBMENTION_QUEUE", "queue3")

	handle := "eventHandle"
	body := "{ \"url\": \"anysource\" }"

	ctx := context.Background()

	html := `<html><body><p><a href="anytarget">Title</a></body></html>`

	mockHTTP := new(MockHTTPClient)
	mockHTTP.On("Do", mock.Anything).Return(mockResponse(html), nil).Once()

	mockSQS := new(MockSQSClient)
	mockSQS.On("SendMessage", ctx, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
		return *input.QueueUrl == "queue2" && *input.MessageBody == body
	}), mock.Anything).Return(nil, nil).Once()
	mockSQS.On("SendMessage", ctx, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
		return *input.QueueUrl == "queue3" && matches("anysource", "anytarget", *input.MessageBody)
	}), mock.Anything).Return(nil, nil).Once()
	mockSQS.On("DeleteMessage", ctx, mock.MatchedBy(func(input *sqs.DeleteMessageInput) bool {
		return *input.QueueUrl == "queue1" && *input.ReceiptHandle == handle
	}), mock.Anything).Return(nil, nil).Once()

	var x = MicroService{
		SQSC:  mockSQS,
		HTTPC: mockHTTP,
	}

	err := x.processMessage(ctx, handle, body)
	assert.NoError(t, err)

	mockSQS.AssertExpectations(t)
}
