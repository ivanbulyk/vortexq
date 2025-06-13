package broker

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

// Test Publish and Topics map behavior
func TestPublishTopics(t *testing.T) {
	v := NewVortexQ[string]()
	msg := Message[string]{ID: "1", Pattern: "topic1", Data: "hello"}

	// Publish first message
	v.Publish(msg)
	val, ok := v.Topics.Load("topic1")
	if !ok {
		t.Fatalf("expected topic1 in Topics map")
	}
	msgs, ok := val.([]Message[string])
	if !ok {
		t.Fatalf("invalid type for messages: %T", val)
	}
	if got, want := len(msgs), 1; got != want {
		t.Fatalf("got %d messages, want %d", got, want)
	}

	// Publish second message to same topic
	msg2 := Message[string]{ID: "2", Pattern: "topic1", Data: "world"}
	v.Publish(msg2)
	val2, ok := v.Topics.Load("topic1")
	if !ok {
		t.Fatalf("expected topic1 in Topics map after second publish")
	}
	msgs2 := val2.([]Message[string])
	if got, want := len(msgs2), 2; got != want {
		t.Fatalf("got %d messages, want %d", got, want)
	}
}

// Test Subscribe and Subscriptions map behavior
func TestSubscribeTopics(t *testing.T) {
	v := NewVortexQ[string]()
	s1 := Subscription{ID: "s1", SubscriberAddress: "addr1", TopicName: "t1"}
	s2 := Subscription{ID: "s2", SubscriberAddress: "addr2", TopicName: "t1"}

	// Subscribe first
	if err := v.Subscribe(s1); err != nil {
		t.Fatalf("unexpected error on Subscribe: %v", err)
	}
	val, ok := v.Subscriptions.Load("t1")
	if !ok {
		t.Fatalf("expected t1 in Subscriptions map")
	}
	subs, ok := val.([]Subscription)
	if !ok {
		t.Fatalf("invalid type for subs: %T", val)
	}
	if got, want := len(subs), 1; got != want {
		t.Fatalf("got %d subs, want %d", got, want)
	}

	// Subscribe second
	if err := v.Subscribe(s2); err != nil {
		t.Fatalf("unexpected error on second Subscribe: %v", err)
	}
	val2, ok := v.Subscriptions.Load("t1")
	if !ok {
		t.Fatalf("expected t1 in Subscriptions map after second Subscribe")
	}
	subs2 := val2.([]Subscription)
	if got, want := len(subs2), 2; got != want {
		t.Fatalf("got %d subs, want %d", got, want)
	}
}

// Test sendWebhook success and payload structure
func TestSendWebhookSuccess(t *testing.T) {
	// Prepare test server to capture the request
	var bufMu sync.Mutex
	var payload WebhookRequest[string]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed reading request body: %v", err)
		}
		bufMu.Lock()
		err = json.Unmarshal(data, &payload)
		bufMu.Unlock()
		if err != nil {
			t.Errorf("failed unmarshaling payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	v := NewVortexQ[string]()
	msg := Message[string]{ID: "1", Pattern: "evt", Data: "d"}
	// sendWebhook should succeed
	if err := v.sendWebhook(msg, server.URL); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate captured payload
	bufMu.Lock()
	got := payload.EventData
	bufMu.Unlock()
	if !reflect.DeepEqual(got, msg) {
		t.Fatalf("got payload %v, want %v", got, msg)
	}
}

// Test sendWebhook failure on non-200 response
func TestSendWebhookNonOK(t *testing.T) {
	// server returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	v := NewVortexQ[string]()
	msg := Message[string]{ID: "x", Pattern: "p", Data: "d"}
	err := v.sendWebhook(msg, server.URL)
	if err == nil {
		t.Fatal("expected error on non-200 response")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("status")) {
		t.Errorf("unexpected error, want status error, got %v", err)
	}
}

// Test Swirl delivers published messages to subscribers and clears topics
func TestSwirlDelivery(t *testing.T) {
	// Prepare test server to capture requests concurrently
	var mu sync.Mutex
	received := make([]Message[string], 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req WebhookRequest[string]
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode error: %v", err)
		}
		mu.Lock()
		received = append(received, req.EventData)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	v := NewVortexQ[string]()
	// Subscribe to topic "topic"
	sub := Subscription{ID: "sub", SubscriberAddress: server.URL, TopicName: "topic"}
	if err := v.Subscribe(sub); err != nil {
		t.Fatalf("subscribe error: %v", err)
	}

	// Publish messages
	msg1 := Message[string]{ID: "1", Pattern: "topic", Data: "a"}
	msg2 := Message[string]{ID: "2", Pattern: "topic", Data: "b"}
	v.Publish(msg1)
	v.Publish(msg2)

	// Swirl should deliver both messages
	if err := v.Swirl(); err != nil {
		t.Fatalf("Swirl error: %v", err)
	}

	// Allow a short moment for handlers to record
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	got := received
	mu.Unlock()
	want := []Message[string]{msg1, msg2}
	if len(got) != len(want) {
		t.Fatalf("received %v messages, want %v messages", len(got), len(want))
	}
	// compare ignoring order
	gotMap := make(map[string]Message[string], len(got))
	for _, m := range got {
		gotMap[m.ID] = m
	}
	for _, w := range want {
		if gm, ok := gotMap[w.ID]; !ok {
			t.Errorf("missing message with ID %q", w.ID)
		} else if gm != w {
			t.Errorf("message mismatch for ID %q: got %v, want %v", w.ID, gm, w)
		}
	}

	// Topics should be cleared
	val, ok := v.Topics.Load("topic")
	if !ok {
		t.Fatalf("expected topic key after Swirl")
	}
	cleared, _ := val.([]Message[string])
	if len(cleared) != 0 {
		t.Fatalf("expected cleared messages, got %v", cleared)
	}
}
