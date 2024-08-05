package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// This microservice receives messages from an SQS queue
// when new posts are available and dispatches messages
// to a notification queue and webmention queues.

type MicroService struct {
	HTTPClient httpDoer
	SNSClient  snsPublisher
}

type Message struct {
	Url string `json:"url"`
}

type Mention struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

const WebMentionEnvVar = "WEBMENTION_TOPIC"

type snsPublisher interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// handle receiving each message
func (s *MicroService) HandleMessage(ctx context.Context, body string) {
	if err := s.processMessage(ctx, body); err != nil {
		log.Fatalf("unable to process message, %v", err)
	}
}

func (s *MicroService) processMessage(ctx context.Context, body string) error {
	var webMentionTopicArn = os.Getenv(WebMentionEnvVar)
	if webMentionTopicArn == "" {
		log.Fatal("unable to get ", WebMentionEnvVar)
	}
	log.Printf("received incoming message: %s", body)

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
		fmt.Printf("found outgoing link: %s\n", link)
		var payload = Mention{
			Source: msg.Url,
			Target: link,
		}
		if err := s.sendMessage(ctx, webMentionTopicArn, toJSON(payload)); err != nil {
			return err
		}
	}

	return nil
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
	req.Header.Set("User-Agent", "Spirit Riot Dispatcher (+https://github.com/tkrehbiel/spiritriot-dispatcher-service)")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	return s.extractLinks(resp.Body)
}

func (s *MicroService) extractLinks(reader io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	// find every link inside an <article> block
	var links []string
	doc.Find("article").Find("a").Each(func(index int, element *goquery.Selection) {
		href, _ := element.Attr("href")
		if s.validateLink(href) {
			links = append(links, href)
		}
	})

	return links, nil
}

func (s *MicroService) validateLink(href string) bool {
	if strings.HasPrefix(href, "http://") {
		// don't mention http links
		return false
	}
	if href[0] == '/' || href[0] == '#' {
		// ignore relative links and fragments
		return false
	}
	u, err := url.Parse(href)
	if err != nil {
		// mangled url
		return false
	}
	if u.Scheme != "https" {
		// only accept https
		return false
	}
	return true
}

func (s *MicroService) sendMessage(ctx context.Context, topicArn string, body string) error {
	log.Printf("sending outgoing message to %s", topicArn)
	_, err := s.SNSClient.Publish(ctx, &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &body,
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
