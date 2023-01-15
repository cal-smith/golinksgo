package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cal-smith/golinks/links"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var tables string

func GetDb(ctx context.Context) *links.Queries {
	db, err := sql.Open("sqlite3", "links.db")
	if err != nil {
		panic(err)
	}

	if _, err := db.ExecContext(ctx, tables); err != nil {
		panic(err)
	}

	return links.New(db)
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	path := html.EscapeString(r.URL.Path)
	path = strings.ReplaceAll(strings.ToLower(path), "-", "")

	log.Println(path, r.RemoteAddr, r.UserAgent())

	ctx := context.Background()
	queries := GetDb(ctx)

	match, err := queries.ExactMatch(ctx, path)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Database error", http.StatusInternalServerError)
		panic(err)
	}

	if match.Source == path {
		log.Println("exact match", path, match.Source)
		http.Redirect(w, r, match.Destination, http.StatusFound)
		return
	} else {
		matches, err := queries.FuzzyMatch(ctx, links.FuzzyMatchParams{Column1: path, Column2: path})
		if err == nil {
			for _, match := range matches {
				log.Println("matching", path, match.Source)
				var (
					s string
				)
				n, err := fmt.Sscanf(path, match.Source, &s)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Sscanf: %v\n", err)
				}
				log.Println(path, "param:", s)
				log.Println(path, "matched:", n)
				if n > 0 {
					dest := match.Destination
					if strings.Contains(dest, "%s") {
						dest = fmt.Sprintf(match.Destination, s)
					}
					log.Println(path, "redirecting to", dest)
					http.Redirect(w, r, dest, http.StatusFound)
					return
				}
			}
		}
	}

	link_list, err := queries.Listlinks(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(err)
	}

	index, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}

	data := struct {
		Current string
		Links   []links.ListlinksRow
	}{
		Current: path,
		Links:   link_list,
	}

	err = index.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func apiGoLinkHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path, r.RemoteAddr, r.UserAgent())

	ctx := context.Background()
	queries := GetDb(ctx)

	source := r.PostFormValue("source")
	destination := r.PostFormValue("destination")
	description := sql.NullString{
		String: r.PostFormValue("description"),
		Valid:  true,
	}

	if strings.TrimSpace(source) == "" || strings.TrimSpace(destination) == "" {
		http.Error(w, "Invalid golink: no source or destination", http.StatusBadRequest)
		return
	}

	_, err := queries.Createlink(ctx, links.CreatelinkParams{Source: source, Destination: destination, Description: description})
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func listTop(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	queries := GetDb(ctx)

	list, err := queries.ListTop(ctx, 10)
	if err != nil {
		panic(err)
	}

	jsonList, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonList)
}

func listSearch(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	queries := GetDb(ctx)

	searchTerm := r.URL.Query().Get("q")
	if strings.TrimSpace(searchTerm) == "" {
		http.Error(w, "Invalid search query: no query provided", http.StatusBadRequest)
		return
	}

	list, err := queries.Search(ctx, searchTerm, 10)
	if err != nil {
		panic(err)
	}

	jsonList, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonList)
}

func main() {
	fmt.Println("Hello, World! yeet on 8000")

	logger := log.Default()

	requestTimer := promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "golinks_requests_duration",
			Help:    "Request duration",
			Buckets: prometheus.LinearBuckets(10.0, 15.0, 10),
		},
	)

	requestCounter := promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "golinks_requests_total",
			Help: "The total number of requests",
		},
	)

	http.Handle("/", Adapt(
		http.HandlerFunc(goHandler),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "GO")))

	http.Handle("/api/golink", Adapt(
		http.HandlerFunc(apiGoLinkHandler),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	http.Handle("/api/top", Adapt(
		http.HandlerFunc(listTop),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	http.Handle("/api/search", Adapt(
		http.HandlerFunc(listSearch),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	// Expose /metrics HTTP endpoint
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8000", nil))
}
