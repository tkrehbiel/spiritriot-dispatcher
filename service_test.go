package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcess(t *testing.T) {
	arn := "topicArn"
	os.Setenv("WEBMENTION_TOPIC", arn)

	body := "{ \"url\": \"anysource\" }"

	ctx := context.Background()

	html := `<html><body><article><p><a href="https://anytarget">Title</a><article></body></html>`

	mockHTTP := new(MockHTTPClient)
	mockHTTP.On("Do", mock.Anything).Return(mockResponse(html), nil).Once()

	mockSNS := new(MockSNSClient)
	mockSNS.On("Publish", ctx, mock.MatchedBy(func(input *sns.PublishInput) bool {
		return *input.TopicArn == arn && matches("anysource", "https://anytarget", *input.Message)
	}), mock.Anything).Return(nil, nil).Once()

	var svc = MicroService{
		SNSClient:  mockSNS,
		HTTPClient: mockHTTP,
	}

	err := svc.processMessage(ctx, body)
	assert.NoError(t, err)

	mockSNS.AssertExpectations(t)
}

func matches(source string, target string, js string) bool {
	var mention Mention
	if err := json.Unmarshal([]byte(js), &mention); err != nil {
		return false
	}
	return source == mention.Source && target == mention.Target
}

func TestExtractLinks_NoArticle(t *testing.T) {
	html := `<html><body><a href="https://anytarget">Title</a></body></html>`

	var svc MicroService
	links, err := svc.extractLinks(strings.NewReader(html))

	assert.NoError(t, err)
	assert.Equal(t, 0, len(links))
}

func TestExtractLinks_MultipleArticles(t *testing.T) {
	html := `<html><body><article><a href="https://url1">Title</a></article><article><a href="https://url2">Title</a></article></body></html>`

	var svc MicroService
	links, err := svc.extractLinks(strings.NewReader(html))

	assert.NoError(t, err)
	assert.Equal(t, 2, len(links))
	assert.Equal(t, []string{"https://url1", "https://url2"}, links)
}

func TestExtractLinks_IgnoreFragments(t *testing.T) {
	html := `<html><body><article><a href="#fragment1">Title</a><a href="#fragment2">Title</a></article></body></html>`

	var svc MicroService
	links, err := svc.extractLinks(strings.NewReader(html))

	assert.NoError(t, err)
	assert.Equal(t, 0, len(links))
}

func TestExtractLinks_IgnoreRelative(t *testing.T) {
	html := `<html><body><article><a href="/relativepath1">Title</a><a href="/notsure/why/I/need/totest/two/relativepath2">Title</a></article></body></html>`

	var svc MicroService
	links, err := svc.extractLinks(strings.NewReader(html))

	assert.NoError(t, err)
	assert.Equal(t, 0, len(links))
}

func TestExtractLinks_IgnoreMangled(t *testing.T) {
	html := `<html><body><article><a href="sftp://ugh/no/ftp">Title</a><a href="mailto://why/would/anyone/use/mail/links">Title</a><a href="alskjdf://@#$asdf"></a></article></body></html>`

	var svc MicroService
	links, err := svc.extractLinks(strings.NewReader(html))

	assert.NoError(t, err)
	assert.Equal(t, 0, len(links))
}
