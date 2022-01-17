package common

import (
	"encoding/json"
)

// EventType represents the current type of event. e.g.: run.initialize, run.finished, and so on.
type EventType string

const (
	// RunInitialize represents Jenkins run is initializing.
	RunInitialize EventType = "run.initialize"
	// RunStarted represents Jenkins run has started.
	RunStarted EventType = "run.started"
	// RunFinalized represents Jenkins run has finalized.
	RunFinalized EventType = "run.finalized"
	// RunCompleted represents Jenkins run has completed.
	RunCompleted EventType = "run.completed"
	// RunDeleted represents Jenkins run has been deleted.
	RunDeleted EventType = "run.deleted"
)

// Event contains common fields of event except event data.
type Event struct {
	Type     string          `json:"type"`
	Source   string          `json:"source"`
	ID       string          `json:"id"`
	Time     string          `json:"time"`
	DataType string          `json:"dataType"`
	Data     json.RawMessage `json:"data"`
}

// TypeEquals checks the given type is equal to type in event.
// Return true if event is not nil and types are equal, false otherwise.
func (event *Event) TypeEquals(eventType EventType) bool {
	return event != nil && event.Type == string(eventType)
}
