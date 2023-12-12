package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// This microservice simply receives a message from one
// SQS queue and dispatches messages to other SQS queues.

type MicroService struct {
	HTTPC httpClient
	SQSC  sqsClient
}

type Message struct {
	Url string `json:"url"`
}

type Mention struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// SQS client interface to allow mocking
type sqsClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// handle receiving each SQS message
func (s *MicroService) HandleMessage(ctx context.Context, handle string, body string) {
	if err := s.processMessage(ctx, handle, body); err != nil {
		log.Fatalf("unable to process message, %v", err)
	}
}

func (s *MicroService) processMessage(ctx context.Context, handle string, body string) error {
	var notifierQueue = os.Getenv("NOTIFIER_QUEUE")
	if notifierQueue == "" {
		log.Fatal("unable to get NOTIFIER_QUEUE")
	}
	var webmentionQueue = os.Getenv("WEBMENTION_QUEUE")
	if webmentionQueue == "" {
		log.Fatal("unable to get WEBMENTION_QUEUE")
	}
	log.Printf("received incoming message %s: %s", handle, body)

	// queue a notification about the post
	if err := s.sendMessage(ctx, notifierQueue, body); err != nil {
		return err
	}

	// scan the post for outgoing links
	var msg Message
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return err
	}
	links, err := s.scanPost(ctx, msg.Url)
	if err != nil {
		return err
	}

	// queue a webmention for each link
	for _, link := range links {
		var payload = Mention{
			Source: msg.Url,
			Target: link,
		}
		if err := s.sendMessage(ctx, webmentionQueue, toJSON(payload)); err != nil {
			return err
		}
	}

	// delete incoming message after handling it
	return s.deleteMessage(ctx, handle)
}

func (s *MicroService) scanPost(ctx context.Context, url string) ([]string, error) {
	// Originally I only planned for this service to
	// dispatch messages based on incoming messages,
	// but it makes sense to scan the HTML content
	// of the new page to see if we need to queue
	// webmention outputs.
	timeout, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	req, err := http.NewRequestWithContext(timeout, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.HTTPC.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var links []string
	doc.Find("a").Each(func(index int, element *goquery.Selection) {
		href, _ := element.Attr("href")
		text := element.Text()
		fmt.Printf("found outgoing link: [%s](%s)\n", text, href)
		links = append(links, href)
	})

	return links, nil
}

func (s *MicroService) sendMessage(ctx context.Context, queue string, body string) error {
	log.Printf("sending outgoing message to %s", queue)
	_, err := s.SQSC.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queue,
		MessageBody: &body,
	})
	return err
}

// delete SQS message after it's handled
func (s *MicroService) deleteMessage(ctx context.Context, handle string) error {
	var incomingQueue = os.Getenv("INCOMING_QUEUE")
	if incomingQueue == "" {
		log.Fatal("unable to get INCOMING_QUEUE")
	}
	log.Printf("deleting incoming message %s", handle)
	_, err := s.SQSC.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &incomingQueue,
		ReceiptHandle: &handle,
	})
	return err
}

func toJSON(v any) string {
	serialized, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(serialized)
}
