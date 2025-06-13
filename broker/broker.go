package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ivanbulyk/vortexq/internal/logging"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type VortexQ[T any] struct {
	Subscriptions sync.Map     `json:"subscriptions"`
	Topics        sync.Map     `json:"topics"`
	Logger        *slog.Logger `json:"-"`
}

func NewVortexQ[T any]() *VortexQ[T] {
	return &VortexQ[T]{
		Subscriptions: sync.Map{},
		Topics:        sync.Map{},
		Logger:        slog.Default(),
	}
}

type VortexQFuncs interface {
	Publish(message Message[any])
	Subscribe(subscription Subscription) error
	sendWebhook(message Message[any], subscriberAddress string) error
	Swirl() error
}

type WebhookRequest[T any] struct {
	EventType string     `json:"event_type"`
	EventData Message[T] `json:"event_data"`
	Timestamp time.Time  `json:"timestamp"`
}

type Message[T any] struct {
	ID      string `json:"id"`
	Pattern string `json:"pattern"`
	Data    T      `json:"data"`
}

type Subscription struct {
	ID                string `json:"id"`
	SubscriberAddress string `json:"subscriber_address"`
	TopicName         string `json:"topic_name"`
}

func (vq *VortexQ[T]) Swirl() error {
	const op = "broker.VortexQ.Swirl"
	var wg sync.WaitGroup

	vq.Subscriptions.Range(func(key, value interface{}) bool {
		topicKey := key
		subs := value.([]Subscription)

		topicVal, ok := vq.Topics.Load(topicKey)
		if !ok {
			vq.Logger.With(slog.String("op", op)).
				Info("topic can't be found", logging.Attr("topic", topicKey))
			return true
		}

		messages := topicVal.([]Message[T])

		for _, msg := range messages {
			msgCopy := msg // capture by value
			for _, sub := range subs {
				subCopy := sub // capture by value
				wg.Add(1)
				go func() {
					defer wg.Done()
					vq.Logger.With(slog.String("op", op)).
						Info("sending message", logging.Attr("message", msgCopy),
							"to subscriber", logging.Attr("subscriber", subCopy.SubscriberAddress))
					err := vq.sendWebhook(msgCopy, subCopy.SubscriberAddress)
					if err != nil {
						vq.Logger.With(slog.String("op", op)).Error("error sending webhook to",
							subCopy.SubscriberAddress, logging.Err(err))
					}
				}()
			}
		}
		// clear messages so we don't redeliver next time
		vq.Topics.Store(topicKey, []Message[T]{})
		return true
	})
	wg.Wait()
	return nil
}

func (vq *VortexQ[T]) Subscribe(subscription Subscription) error {
	const op = "broker.VortexQ.Subscribe"
	// if the topic exists, add the subscription to the topic
	if topicName, ok := vq.Subscriptions.Load(subscription.TopicName); !ok {
		subs := make([]Subscription, 0)
		subs = append(subs, subscription)
		vq.Subscriptions.Store(subscription.TopicName, subs)
		vq.Logger.With(slog.String("op", op)).
			Info("new subscription created for topic:", logging.Attr("topic", subscription.TopicName))
	} else {
		subs := append(topicName.([]Subscription), subscription)
		vq.Subscriptions.Store(subscription.TopicName, subs)
	}

	return nil
}

func (vq *VortexQ[T]) Publish(msg Message[T]) {
	const op = "broker.VortexQ.Publish"
	if topicRaw, ok := vq.Topics.Load(msg.Pattern); ok {
		msgs := topicRaw.([]Message[T])
		msgs = append(msgs, msg)
		vq.Topics.Store(msg.Pattern, msgs)
	} else {
		newTopic := []Message[T]{msg}
		vq.Topics.Store(msg.Pattern, newTopic)
		vq.Logger.With(slog.String("op", op)).
			Info("new topic created", logging.Attr("topic", msg.Pattern))
	}

}

func (vq *VortexQ[T]) sendWebhook(msg Message[T], SubscriberAddr string) error {
	const op = "broker.VortexQ.SendWebhook"
	// create a Webhook payload
	wreq := WebhookRequest[T]{
		EventType: msg.Pattern,
		EventData: msg,
		Timestamp: time.Now().UTC(),
	}

	// Marshal the WebhookRequest to JSON
	jsonBytes, err := json.Marshal(wreq)
	if err != nil {
		return fmt.Errorf("error creating webhook payload: %w", err)
	}

	// Prepare the webhook request
	req, err := http.NewRequest("POST", SubscriberAddr, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to prepare the webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the webhook to the callback URL
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending webhook to %s: %w", SubscriberAddr, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			vq.Logger.With(slog.String("op", op)).Error("error closing response body:", logging.Err(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook delivery failed: status %s", resp.Status)
	}

	vq.Logger.With(slog.String("op", op)).
		Info("webhook delivered to", logging.Attr("subscriber address", SubscriberAddr),
			logging.Attr("with status", resp.Status))
	return nil
}
