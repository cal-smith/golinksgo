package main

import (
	"context"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cal-smith/golinks/links"
	"github.com/prometheus/client_golang/prometheus"
)

type Adapter func(http.Handler) http.Handler

// inspired by https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

func CanonicalLogLine(logger *log.Logger, scope string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := html.EscapeString(r.URL.Path)
			h.ServeHTTP(w, r)
			logger.Printf("CANONICAL-%s-LINE path=%s duration=%d address=%s user_agent=%s",
				strings.ToUpper(scope),
				path,
				time.Since(start),
				r.RemoteAddr,
				r.UserAgent())
		})
	}
}

func trackView(requestCounter prometheus.Counter, ctx context.Context, p links.CreatePageViewParams) {
	requestCounter.Inc()
	queries := GetDb(ctx)
	_, err := queries.CreatePageView(ctx, p)
	if err != nil {
		log.Println("error tracking pageview")
	}
}

func RequestMetrics(requestTimer prometheus.Histogram, requestCounter prometheus.Counter) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := html.EscapeString(r.URL.Path)
			ctx := context.Background()
			trackView(requestCounter, ctx, links.CreatePageViewParams{Path: path, Ip: r.RemoteAddr})
			h.ServeHTTP(w, r)
			requestTimer.Observe(float64(time.Since(start)))
		})
	}
}
