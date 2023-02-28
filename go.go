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
	"github.com/julienschmidt/httprouter"
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
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	log.Println(path, r.RemoteAddr, r.UserAgent(), r.Method)

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
		matches, err := queries.FuzzyMatch(ctx, links.FuzzyMatchParams{Source: path, Column2: path})
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

	for i, link := range link_list {
		if link.Source[0] != '/' {
			link_list[i].Source = "/" + link.Source
		}
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

	if path == "/" {
		data.Current = ""
	}

	err = index.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func apiDeleteGoLinkHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path, r.RemoteAddr, r.UserAgent(), r.Method)
	params := httprouter.ParamsFromContext(r.Context())
	ctx := context.Background()
	queries := GetDb(ctx)

	link := "/" + params.ByName("link")

	err := queries.DeleteLink(ctx, link)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting link %s", link), http.StatusInternalServerError)
		panic(err)
	}

	err = queries.DeleteView(ctx, link)
	if err != nil {
		log.Println("error cleaning up link views for: ", link)
		log.Println(err)
	}

	w.Header().Add("Redirect", "/")
	w.Write([]byte("ok"))
}

var ErrEmptyLink = errors.New("no source or destination")

func verifyCreateOrUpdateLink(r *http.Request) (links.CreatelinkParams, error) {
	source := r.PostFormValue("source")
	if source[0] != '/' {
		source = "/" + source
	}

	destination := r.PostFormValue("destination")
	description := sql.NullString{
		String: r.PostFormValue("description"),
		Valid:  true,
	}

	if strings.TrimSpace(source) == "" || strings.TrimSpace(destination) == "" {
		return links.CreatelinkParams{}, ErrEmptyLink
	}

	return links.CreatelinkParams{
		Source:      source,
		Destination: destination,
		Description: description,
	}, nil
}

func apiPutGoLinkHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path, r.RemoteAddr, r.UserAgent(), r.Method)
	params := httprouter.ParamsFromContext((r.Context()))
	ctx := context.Background()
	queries := GetDb(ctx)

	updateParams, err := verifyCreateOrUpdateLink(r)

	if err == ErrEmptyLink {
		http.Error(w, "Invalid golink: no source or destination", http.StatusBadRequest)
		return
	}

	newLink, err := queries.UpdateLink(ctx, links.UpdateLinkParams{
		Source:      updateParams.Source,
		Destination: updateParams.Destination,
		Description: updateParams.Description,
		Source_2:    "/" + params.ByName("link"),
	})
	if err != nil {
		panic(err)
	}

	jsonLink, err := json.Marshal(newLink)
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Redirect", "/")
	w.Write(jsonLink)
}

func apiGoLinkHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path, r.RemoteAddr, r.UserAgent(), r.Method)

	ctx := context.Background()
	queries := GetDb(ctx)

	params, err := verifyCreateOrUpdateLink(r)

	if err == ErrEmptyLink {
		http.Error(w, "Invalid golink: no source or destination", http.StatusBadRequest)
		return
	}

	_, err = queries.Createlink(ctx, params)
	if err != nil {
		log.Println("tried creating a link with:", params.Source, params.Destination, params.Description)
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

	router := httprouter.New()

	router.NotFound = Adapt(
		http.HandlerFunc(goHandler),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "GO"))

	router.Handler(http.MethodPost, "/api/golink", Adapt(
		http.HandlerFunc(apiGoLinkHandler),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	router.Handler(http.MethodDelete, "/api/golink/:link", Adapt(
		http.HandlerFunc(apiDeleteGoLinkHandler),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	router.Handler(http.MethodPut, "/api/golink/:link", Adapt(
		http.HandlerFunc(apiPutGoLinkHandler),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	router.Handler(http.MethodGet, "/api/top", Adapt(
		http.HandlerFunc(listTop),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	router.Handler(http.MethodGet, "/api/search", Adapt(
		http.HandlerFunc(listSearch),
		RequestMetrics(requestTimer, requestCounter),
		CanonicalLogLine(logger, "API")))

	// Expose /metrics HTTP endpoint
	router.Handler(http.MethodGet, "/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8000", router))
}
