package monitor2

import (
	"encoding/json"
	"fmt"
	"html/template"
	"monitor2/src/crawler"
	database "monitor2/src/db"
	"monitor2/src/diffs"
	"monitor2/src/repositories"
	"strconv"
	"time"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type App struct {
	Router    *mux.Router
	DB        *database.Database
	templates *template.Template
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var headers []byte
		var err error

		headers, err = json.Marshal(&r.Header)
		if err != nil {
			log.Print(err)
		}

		log.Info().
			Str("Method", r.Method).
			Str("Host", r.Host).
			Str("URI", r.RequestURI).
			RawJSON("Headers", headers).
			Msg("")

		next.ServeHTTP(w, r)
	})
}

func (app *App) Init(addr string) {
	app.templates = template.Must(template.ParseGlob("static/templates/*.html"))

	app.Router = mux.NewRouter()
	app.Router.Use(loggingMiddleware)
	app.Router.HandleFunc("/", app.HomeHandler)
	app.Router.HandleFunc("/health", app.HealthCheck)

	app.Router.HandleFunc("/crawl", crawler.Endpoints)
	app.Router.HandleFunc("/crawl/c", crawler.CreateEndpoint)
	app.Router.HandleFunc("/crawl/u", crawler.UpdateEndpoint)
	app.Router.HandleFunc("/crawl/d", crawler.DeleteEndpoint)
	app.Router.HandleFunc("/crawl/run", app.RunSchedule)
	app.Router.HandleFunc("/crawl/run_single", crawler.RunEndpoint)

	app.Router.HandleFunc("/repos", repositories.Repos)
	app.Router.HandleFunc("/repos/c", repositories.CreateRepo)
	app.Router.HandleFunc("/repos/u", repositories.UpdateRepo)

	app.Router.HandleFunc("/diffs", diffs.Diffs)
	app.Router.HandleFunc("/diff/{id}", diffs.Diff)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Starting server at: %s", addr)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func (_ App) HomeHandler(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, "static/index.html")
}

func (_ App) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func (app App) RunSchedule(w http.ResponseWriter, r *http.Request) {
	s := r.PostFormValue("s")
	if len(s) == 0 {
		fmt.Fprintf(w, "Missing 's' param")
		return
	}

	schedule, err := strconv.Atoi(s)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	n, errors := crawler.RunBySchedule(schedule, &database.DB)
	if len(errors) != 0 {
		fmt.Fprint(w, errors)
		return
	}

	fmt.Fprintf(w, "Running crawler on %+v endpoint(s).", n)
}
