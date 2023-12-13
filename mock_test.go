package main

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/mock"
)

type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if o, ok := args.Get(0).(*sqs.SendMessageOutput); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if o, ok := args.Get(0).(*sqs.DeleteMessageOutput); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockSNSClient struct {
	mock.Mock
}

func (m *MockSNSClient) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	args := m.Called(ctx, params, optFns)
	if o, ok := args.Get(0).(*sns.PublishOutput); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if o, ok := args.Get(0).(*http.Response); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func mockResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
}
