package live

import (
	"context"
	"fmt"
	"sync"
)

var mu sync.Mutex

type subscriber struct {
	notificationChan chan LiveEvent
}

// New returns a scoped set of Subscriptions. Use the returned subscriptions
// methods Emit and Subscribe to interact.
func New() Subscriptions {
	return make(Subscriptions)
}

// Subscriptions is a map with key for the content type's name, whose value is
// a map of the live event types, each having a value of a slice of subscribers.
type Subscriptions map[string]map[liveEventType][]subscriber

// Emit pushes a notification to all related subscribers within the same event
// type, for the same content type.
func (subs Subscriptions) Emit(ctx context.Context, contentType string, data interface{}, eventType liveEventType) error {
	// iterate through the subscribers for this event and send on their
	// notification channel
	for contentT, eventTMap := range subs {
		if contentT != contentType {
			continue
		}

		subs, ok := eventTMap[eventType]
		if !ok {
			return QueryError(
				fmt.Sprintf("no subs for %s.%d", contentType, eventType),
			)
		}

		for i := range subs {
			subs[i].notify(data, eventType)
		}
	}

	// if error, return QueryError
	return nil
}

func (s subscriber) notify(data interface{}, eventType liveEventType) {
	s.notificationChan <- LiveEvent{
		Type:    eventType,
		content: data,
	}
}

// Subscribe returns a recieve-only channel of ListEvent values. Use its Content
// method to grab the underlying data, and make a type assertion to use it as a
// Ponzu content type.
func (subs Subscriptions) Subscribe(contentType string, eventType liveEventType) <-chan LiveEvent {
	if _, ok := subs[contentType]; !ok {
		mu.Lock()
		subs[contentType] = make(map[liveEventType][]subscriber)
		mu.Unlock()
	}

	notifs := make(chan LiveEvent)

	// add a new subscriber with the nofitication chan and return the chan to
	// the caller to recieve from
	mu.Lock()
	subs[contentType][eventType] = append(subs[contentType][eventType],
		subscriber{
			notificationChan: notifs,
		},
	)
	mu.Unlock()

	return notifs
}

// QueryError is an error provided to specify errors from the live package.
type QueryError string

func (e QueryError) Error() string {
	return fmt.Sprintf("[live.error] %s", e)
}
