package twitter

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Twitter struct {
	client *twitter.Client
	stream *twitter.Stream
	ch     chan *twitter.Tweet
}

func NewTwitter(consumerKey, consumerSecret, accessToken, accessSecret string, ch chan *twitter.Tweet) *Twitter {
	config := oauth1.NewConfig(consumerKey, consumerSecret)

	return &Twitter{
		twitter.NewClient(config.Client(oauth1.NoContext, oauth1.NewToken(accessToken, accessSecret))),
		nil,
		ch,
	}
}

func (t Twitter) Start() error {
	demux := twitter.NewSwitchDemux()
	demux.Tweet = t.handleTweet
	demux.Event = t.handleEvent

	fmt.Println("🐔 Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Follow:        []string{"1016357953807400960"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := t.client.Streams.Filter(filterParams)
	if err == nil {
		t.stream = stream
		go demux.HandleChan(t.stream.Messages)
	}
	return err
}

func (t Twitter) Stop() {
	fmt.Println("🐔 Stopping Stream...")
	if t.stream != nil {
		t.stream.Stop()
		t.stream = nil
	}
}

func (t Twitter) handleTweet(tweet *twitter.Tweet) {
	fmt.Printf("🐔🏃 Processing tweet %v\n", tweet.Text)
	if tweet.User.ID == 1016357953807400960 {
		fmt.Println("🐔🏃 Sending to channel")
		t.ch <- tweet
	} else {
		fmt.Println("🐔🏃 Not worthy")
	}
}

func (t Twitter) handleEvent(event *twitter.Event) {
	fmt.Printf("🐔 %#v\n", event)
}
