package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func newTestRateLimiter(requestsPerMinute int64, burst int64, trustProxyHeaders bool) *rateLimitMiddleware {
	return NewRateLimitMiddleware(requestsPerMinute, burst, trustProxyHeaders).(*rateLimitMiddleware)
}

func TestRateLimitAllowsBurstThenBlocks(t *testing.T) {
	limiter := newTestRateLimiter(60, 3, false)

	current := time.Unix(0, 0)
	limiter.now = func() time.Time { return current }

	for i := range 3 {
		if !limiter.allow("client") {
			t.Fatalf("expected request %d to be allowed within the burst", i+1)
		}
	}

	if limiter.allow("client") {
		t.Error("expected the request after the burst to be blocked")
	}

	// 60 requests per minute is one token per second.
	current = current.Add(time.Second)

	if !limiter.allow("client") {
		t.Error("expected a token to have refilled after a second")
	}
}

func TestRateLimitIsPerClient(t *testing.T) {
	limiter := newTestRateLimiter(60, 1, false)

	current := time.Unix(0, 0)
	limiter.now = func() time.Time { return current }

	if !limiter.allow("first") {
		t.Fatal("expected the first client to be allowed")
	}

	if limiter.allow("first") {
		t.Error("expected the first client to be blocked after its burst")
	}

	if !limiter.allow("second") {
		t.Error("expected a different client to have its own budget")
	}
}

func TestRateLimitDoesNotRefillPastBurst(t *testing.T) {
	limiter := newTestRateLimiter(60, 2, false)

	current := time.Unix(0, 0)
	limiter.now = func() time.Time { return current }

	limiter.allow("client")

	// Idle for far longer than it takes to refill.
	current = current.Add(time.Hour)

	for i := range 2 {
		if !limiter.allow("client") {
			t.Fatalf("expected a full bucket after a long idle period, but request %d was blocked", i+1)
		}
	}

	if limiter.allow("client") {
		t.Error("expected the bucket to be capped at the burst size")
	}
}

func TestRateLimitMiddlewareReturnsTooManyRequests(t *testing.T) {
	limiter := newTestRateLimiter(60, 1, false)

	current := time.Unix(0, 0)
	limiter.now = func() time.Time { return current }

	handler := limiter.MiddlewareFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodPost, "/data/poi", nil)
	request.RemoteAddr = "192.0.2.1:1234"

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, request)

	if first.Code != http.StatusOK {
		t.Fatalf("expected the first request to succeed, but got %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, request)

	if second.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, but got %d", second.Code)
	}

	if second.Header().Get("Retry-After") == "" {
		t.Error("expected a Retry-After header on a rate limited response")
	}
}

func TestClientKey(t *testing.T) {
	tests := []struct {
		name              string
		trustProxyHeaders bool
		remoteAddr        string
		forwardedFor      string
		expected          string
	}{
		{"remote address", false, "192.0.2.1:1234", "", "192.0.2.1"},
		{"proxy header ignored when untrusted", false, "192.0.2.1:1234", "203.0.113.9", "192.0.2.1"},
		{"proxy header used when trusted", true, "192.0.2.1:1234", "203.0.113.9", "203.0.113.9"},
		{"first proxy entry wins", true, "192.0.2.1:1234", "203.0.113.9, 198.51.100.2", "203.0.113.9"},
		{"empty proxy header falls back", true, "192.0.2.1:1234", "", "192.0.2.1"},
		{"address without port", false, "192.0.2.1", "", "192.0.2.1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limiter := newTestRateLimiter(60, 1, test.trustProxyHeaders)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = test.remoteAddr
			if test.forwardedFor != "" {
				request.Header.Set("X-Forwarded-For", test.forwardedFor)
			}

			if actual := limiter.clientKey(request); actual != test.expected {
				t.Errorf("expected key %q, but got %q", test.expected, actual)
			}
		})
	}
}

// The limiter is shared by every request, so its bookkeeping must be safe under
// concurrent access. Run with -race.
func TestRateLimitIsConcurrencySafe(t *testing.T) {
	limiter := newTestRateLimiter(6000, 100, false)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			for range 20 {
				limiter.allow("client")
				limiter.allow(string(rune('a' + i%26)))
			}
		})
	}
	wg.Wait()
}
