package main

import "fmt"

type Event struct {
	Type    string
	Payload interface{} // Data associated with the event
}

// EventHandler is a function type that handles events.
type EventHandler func(event Event)

// EventEmitter manages event handlers and dispatches events.
type EventEmitter struct {
	handlers map[string][]EventHandler // Maps event types to a slice of handlers
}

// NewEventEmitter creates and returns a new EventEmitter instance.
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		handlers: make(map[string][]EventHandler),
	}
}

// On registers an event handler for a specific event type.
func (em *EventEmitter) On(eventType string, handler EventHandler) {
	em.handlers[eventType] = append(em.handlers[eventType], handler)
	fmt.Printf("Registered handler for event type: %s\n", eventType)
}

// Emit dispatches an event to all registered handlers for its type.
// Handlers are executed synchronously (one after another).
func (em *EventEmitter) Emit(event Event) {
	handlers, found := em.handlers[event.Type]
	if !found || len(handlers) == 0 {
		fmt.Printf("No handlers registered for event type: %s\n", event.Type)
		return
	}

	fmt.Printf("Emitting event of type: %s with payload: %v\n", event.Type, event.Payload)

	// Execute each handler synchronously.
	for _, handler := range handlers {
		handler(event)
	}
}
