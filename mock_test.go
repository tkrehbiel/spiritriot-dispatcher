package main

import (
	"context"

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
