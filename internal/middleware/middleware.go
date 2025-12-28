package middleware

import (
	"log"
	"net/http"
	"net/url"
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
