package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivanbulyk/vortexq/broker"
	"github.com/ivanbulyk/vortexq/internal/version"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// helper to perform a request on gin engine
func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestIndexHandler(t *testing.T) {
	h := NewVortexQHandler(nil)
	info := &version.Version{Project: "p", BuildTime: "bt", Commit: "c", Release: "r"}
	h.Version = info

	r := gin.New()
	r.GET("/", h.IndexHandler)

	w := performRequest(r, http.MethodGet, "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("IndexHandler status = %d; want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Message string          `json:"message"`
		Info    version.Version `json:"info"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
	if resp.Info != *info {
		t.Errorf("info = %+v; want %+v", resp.Info, *info)
	}
}

// TestPublishHandler verifies publishing via HTTP updates the broker topics
func TestPublishHandler(t *testing.T) {
	// use real broker as VortexQFuncs implementation
	vq := broker.NewVortexQ[any]()
	h := NewVortexQHandler(vq)
	r := gin.New()
	r.POST("/publish", h.PublishHandler)

	// valid request
	msg := broker.Message[any]{ID: "1", Pattern: "p", Data: "d"}
	body, _ := json.Marshal(msg)
	w := performRequest(r, http.MethodPost, "/publish", bytes.NewReader(body))
	if w.Code != http.StatusOK {
		t.Fatalf("PublishHandler status = %d; want %d", w.Code, http.StatusOK)
	}

	// ensure broker recorded the message
	val, ok := vq.Topics.Load("p")
	if !ok {
		t.Fatalf("expected topic %q", "p")
	}
	msgs := val.([]broker.Message[any])
	if len(msgs) != 1 || !reflect.DeepEqual(msgs[0], msg) {
		t.Errorf("got published messages %v, want [%v]", msgs, msg)
	}

	// invalid JSON
	w = performRequest(r, http.MethodPost, "/publish", bytes.NewReader([]byte("bad")))
	if w.Code != http.StatusBadRequest {
		t.Errorf("PublishHandler bad JSON status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSubscribeHandler verifies subscribing via HTTP updates the broker subscriptions
func TestSubscribeHandler(t *testing.T) {
	vq := broker.NewVortexQ[any]()
	h := NewVortexQHandler(vq)
	r := gin.New()
	r.POST("/subscribe", h.SubscribeHandler)

	sub := broker.Subscription{ID: "1", SubscriberAddress: "a", TopicName: "t"}
	body, _ := json.Marshal(sub)
	w := performRequest(r, http.MethodPost, "/subscribe", bytes.NewReader(body))
	if w.Code != http.StatusOK {
		t.Fatalf("SubscribeHandler status = %d; want %d", w.Code, http.StatusOK)
	}

	// ensure broker recorded the subscription
	val, ok := vq.Subscriptions.Load("t")
	if !ok {
		t.Fatalf("expected subscription for topic %q", "t")
	}
	subs := val.([]broker.Subscription)
	if len(subs) != 1 || subs[0] != sub {
		t.Errorf("got subscriptions %v, want [%v]", subs, sub)
	}

	// invalid JSON
	w = performRequest(r, http.MethodPost, "/subscribe", bytes.NewReader([]byte("bad")))
	if w.Code != http.StatusBadRequest {
		t.Errorf("SubscribeHandler bad JSON status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHealthzAndReadinessHandler(t *testing.T) {
	h := NewVortexQHandler(nil)

	r := gin.New()
	r.GET("/healthz", LivenessHandler)
	r.GET("/readyz", h.ReadinessHandler)

	w := performRequest(r, http.MethodGet, "/healthz", nil)
	if w.Code != http.StatusOK {
		t.Errorf("LivenessHandler status = %d; want %d", w.Code, http.StatusOK)
	}

	w = performRequest(r, http.MethodGet, "/readyz", nil)
	if w.Code != http.StatusOK {
		t.Errorf("ReadinessHandler ready status = %d; want %d", w.Code, http.StatusOK)
	}

	h.IsShuttingDown.Store(true)
	w = performRequest(r, http.MethodGet, "/readyz", nil)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ReadinessHandler shutdown status = %d; want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestPrometheusHandler(t *testing.T) {
	h := NewVortexQHandler(nil)

	// register metrics and seed some values
	h.CustomRegistry.MustRegister(HttpRequestTotal, HttpRequestErrorTotal)
	HttpRequestTotal.WithLabelValues("/foo", "200").Inc()
	HttpRequestErrorTotal.WithLabelValues("/foo", "500").Inc()

	r := gin.New()
	r.GET("/metrics", h.PrometheusHandler())

	w := performRequest(r, http.MethodGet, "/metrics", nil)
	if w.Code != http.StatusOK {
		t.Errorf("PrometheusHandler status = %d; want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Expect seeded metrics in output
	if !strings.Contains(body, "api_http_request_total{path=\"/foo\",status=\"200\"} 1") {
		t.Error("expected api_http_request_total metric in output, got:\n" + body)
	}
	if !strings.Contains(body, "api_http_request_error_total{path=\"/foo\",status=\"500\"} 1") {
		t.Error("expected api_http_request_error_total metric in output, got:\n" + body)
	}
}
