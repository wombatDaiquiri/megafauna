package metrics

import (
	"fmt"
	"github.com/mattn/go-mastodon"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// TODO: Move to separate public Init function.
	prometheus.MustRegister(
		StreamEventUpdate,
	)
}

var (
	StreamEventUpdate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stream_event_update",
			Help: "number of events from given public stream",
		},
		[]string{"stream", "source_stream"},
	)
)

// EventType returns an event type label for metrics.
func EventType(event mastodon.Event) string {
	switch event.(type) {
	case *mastodon.UpdateEvent:
		return "update"
	case *mastodon.DeleteEvent:
		return "delete"
	case *mastodon.NotificationEvent:
		return "notification"
	case *mastodon.ErrorEvent:
		return "error"
	default:
		return fmt.Sprintf("unknown: %T", event)
	}
}
