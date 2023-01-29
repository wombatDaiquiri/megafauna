package exploration

import (
	"context"
	"fmt"
	"github.com/mattn/go-mastodon"
	"github.com/wombatDaiquiri/megafauna/et"
	"net/url"
	"sync"
)

// Office handles stream exploration.
type Office struct {
	ctx          context.Context
	eventHandler func(serverURL string, event mastodon.Event)
	errorHandler func(error)

	lock et.RWLocker
	// TODO: probably holds expeditions
	expeditions map[string]struct{}
}

func NewOffice(ctx context.Context, errHandler func(error), eventHandler func(string, mastodon.Event)) *Office {
	return &Office{
		ctx:          ctx,
		eventHandler: eventHandler,
		errorHandler: errHandler,

		lock:        &sync.RWMutex{},
		expeditions: map[string]struct{}{},
	}
}

func (o *Office) RequestExpedition(serverURL string) {
	if o.expeditionExists(serverURL) {
		return
	}
	o.startExpedition(serverURL)
}

func (o *Office) expeditionExists(serverURL string) bool {
	o.lock.RLock()
	defer o.lock.RUnlock()
	_, ok := o.expeditions[serverURL]
	return ok
}

func (o *Office) startExpedition(serverURL string) {
	o.lock.Lock()
	defer o.lock.Unlock()

	_, alreadyExists := o.expeditions[serverURL]
	if alreadyExists {
		return
	}

	go func() {
		// TODO: emit a metric
		o.explore(serverURL)
		o.lock.Lock()
		delete(o.expeditions, serverURL)
		o.lock.Unlock()
	}()
	o.expeditions[serverURL] = struct{}{}
}

func (o *Office) explore(serverURL string) {
	cli := mastodon.NewClient(&mastodon.Config{Server: serverURL})
	stream, err := cli.StreamingPublic(o.ctx, false)
	if err != nil {
		o.errorHandler(err)
	}
	for {
		select {
		case <-o.ctx.Done():
			return
		case event, ok := <-stream:
			if !ok {
				return
			}
			go o.deriveExpeditions(event)
			o.eventHandler(serverURL, event)
		}
	}
}

func (o *Office) deriveExpeditions(rawEvent mastodon.Event) {
	if rawEvent == nil {
		return
	}
	switch event := rawEvent.(type) {
	case *mastodon.UpdateEvent:
		if event == nil {
			return
		}
		// event.Status.URI
		// https://sportsbots.xyz/users/Blitz_Burgh/statuses/1618752914499182593
		// TODO: run o.RequestExpedition on all servers we can discover
		_ = event.Status
	case *mastodon.NotificationEvent:
		if event == nil {
			return
		}
		// pass for now
	case *mastodon.DeleteEvent:
		if event == nil {
			return
		}
	// pass for now
	default:
		// TODO: proper log
		fmt.Printf("unknown event type: %T\n", event)
	}
}

// ServerFromStatusURI returns a server url from status uri.
func ServerFromStatusURI(statusURI string) (string, error) {
	statusURL, err := url.Parse(statusURI)
	if err != nil {
		return "", fmt.Errorf("[url=%v] parse status URI: %v", statusURI, err)
	}
	serverURL := &url.URL{
		Scheme: statusURL.Scheme,
		Host:   statusURL.Host,
	}
	return serverURL.String(), nil
}
