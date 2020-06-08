package events

import (
	"github.com/gorilla/websocket"
)

// Upgrader upgrades to a ws connection
var Upgrader = websocket.Upgrader{}

type (
	// Event is a websocket generic event
	Event struct {
		Message string
	}

	// EventProcessingService can process events
	EventProcessingService struct {
		events chan Event
	}
)

// EventService is an interface that enables events
type EventService interface {
	processEvent(event Event)
}

// InitialiseEventProcessingService initialises event service
func InitialiseEventProcessingService() *EventProcessingService {
	events := make(chan Event)
	return &EventProcessingService{events: events}
}

// ProcessEvent processes events
func (e *EventProcessingService) ProcessEvent(event Event) {
	e.events <- event
}

// SendEvents sends events
func (e *EventProcessingService) SendEvents(ws *websocket.Conn) {
	for {
		select {
		case event := <-e.events:
			ws.WriteMessage(1, []byte(event.Message))
		}
	}
}
