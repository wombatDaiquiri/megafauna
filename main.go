package main

import (
	"context"
	"fmt"
	"github.com/mattn/go-mastodon"
	"github.com/wombatDaiquiri/megafauna/exploration"
	"github.com/wombatDaiquiri/megafauna/metrics"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO: separate commands for sub-functions
func main() {
	ctx := context.Background()
	// setup scraping
	expeditionOffice := exploration.NewOffice(
		ctx,
		func(err error) { fmt.Printf("error starting stream: %v", err) },
		func(serverURL string, rawEvent mastodon.Event) {
			event, ok := rawEvent.(*mastodon.UpdateEvent)
			if !ok {
				return
			}
			if event == nil {
				return
			}
			sourceServer, err := exploration.ServerFromStatusURI(event.Status.URI)
			if err != nil {
				fmt.Println("ERR GETTING SERVER FROM URI:", err)
			}
			metrics.StreamEventUpdate.WithLabelValues(serverURL, sourceServer).Inc()
		},
	)
	expeditionOffice.RequestExpedition("https://mas.to")

	// setup metrics server
	fmt.Println("setup completed, emiting metrics")
	router := chi.NewRouter()
	router.Get("/metrics", promhttp.Handler().ServeHTTP)
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
