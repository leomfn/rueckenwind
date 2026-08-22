package middleware

import (
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Middleware interface {
	MiddlewareFunc(http.Handler) http.Handler
}

// Logging
type loggingMiddleware struct{}

func NewLoggingMiddleware() Middleware {
	return &loggingMiddleware{}
}

func (m *loggingMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Same site protection
//
// Prevent simple requests to endpoints which are meant to be requested from the
// specified domain only. This is not a real security measure, because this can
// be easily circumvented by spoofing the header. Also, this might be
// problematic for privacy-focused browsers or proxies, that remove the Referer
// header.
//
// TODO: this must be tested in live environment
type sameSiteMiddleware struct {
	debug  bool
	domain string
}

func NewSameSiteMiddleware(domain string, debug bool) Middleware {
	return &sameSiteMiddleware{
		debug:  debug,
		domain: domain,
	}
}

func (m *sameSiteMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.debug {
			next.ServeHTTP(w, r)
			return
		}

		refDomain := m.domain
		if m.debug {
			refDomain = "localhost"
		}

		refHeader := r.Referer()

		requestReferrerURL, err := url.Parse(refHeader)

		if err != nil || requestReferrerURL.Hostname() != refDomain {
			log.Printf("Access to %s blocked, invalid referrer '%s'", r.URL.Path, refHeader)
			http.Error(w, "Invalid Referer", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Rate limiting
//
// Each client gets a token bucket that refills continuously. The point is not
// to stop a determined attacker, but to keep a single misbehaving or looping
// client from burning through the quotas of the upstream APIs, whose usage
// rules the application has to respect.

// Upper bound on the number of clients tracked at once, so that the limiter
// cannot be turned into a memory leak by cycling through source addresses.
const maxTrackedClients = 4096

type bucket struct {
	tokens  float64
	updated time.Time
}

type rateLimitMiddleware struct {
	// Refill rate in tokens per second, and the maximum burst size.
	rate  float64
	burst float64
	// Whether X-Forwarded-For may be believed. Only enable this when the
	// application actually sits behind a reverse proxy that sets it, otherwise
	// clients can trivially spoof their way around the limit.
	trustProxyHeaders bool
	// Overridable so that refills can be tested without sleeping.
	now func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

func NewRateLimitMiddleware(requestsPerMinute int64, burst int64, trustProxyHeaders bool) Middleware {
	return &rateLimitMiddleware{
		rate:              float64(requestsPerMinute) / 60,
		burst:             float64(burst),
		trustProxyHeaders: trustProxyHeaders,
		now:               time.Now,
		buckets:           make(map[string]*bucket),
	}
}

// Identifies the client a request should be accounted to.
func (m *rateLimitMiddleware) clientKey(r *http.Request) string {
	if m.trustProxyHeaders {
		// The left-most entry of X-Forwarded-For is the original client.
		firstEntry, _, _ := strings.Cut(r.Header.Get("X-Forwarded-For"), ",")

		if client := strings.TrimSpace(firstEntry); client != "" {
			return client
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// Takes a token for the given client, reporting whether one was available.
func (m *rateLimitMiddleware) allow(key string) bool {
	now := m.now()

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, found := m.buckets[key]
	if !found {
		if len(m.buckets) >= maxTrackedClients {
			m.evictLocked(now)
		}

		m.buckets[key] = &bucket{tokens: m.burst - 1, updated: now}
		return true
	}

	refill := now.Sub(existing.updated).Seconds() * m.rate
	existing.tokens = math.Min(m.burst, existing.tokens+refill)
	existing.updated = now

	if existing.tokens < 1 {
		return false
	}

	existing.tokens--
	return true
}

// Drops clients whose bucket has fully refilled, since they are indistinguishable
// from clients that were never seen. Callers must hold m.mu.
func (m *rateLimitMiddleware) evictLocked(now time.Time) {
	for key, b := range m.buckets {
		if b.tokens+now.Sub(b.updated).Seconds()*m.rate >= m.burst {
			delete(m.buckets, key)
		}
	}

	// If every tracked client is still active, drop arbitrary ones rather than
	// growing past the bound.
	for key := range m.buckets {
		if len(m.buckets) < maxTrackedClients {
			return
		}

		delete(m.buckets, key)
	}
}

func (m *rateLimitMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.allow(m.clientKey(r)) {
			log.Printf("Rate limit exceeded for %s on %s", m.clientKey(r), r.URL.Path)

			// Suggest waiting for roughly one token to become available.
			w.Header().Set("Retry-After", strconv.FormatInt(int64(math.Ceil(1/m.rate)), 10))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
