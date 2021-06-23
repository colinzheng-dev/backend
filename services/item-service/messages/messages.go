package messages

// ItemEventType is the event type published on the item service's
// "item-change" notification topic.
type ItemEventType string

// Item event types for item creation, update and deletion.
const (
	ItemCreated ItemEventType = "CREATE"
	ItemUpdated ItemEventType = "UPDATE"
	ItemDeleted ItemEventType = "DELETE"
)

// ItemEvent is the message structure published on the item service's
// "item-change" notification topic.
type ItemEvent struct {
	EventType ItemEventType `json:"type"`
	ItemID    string        `json:"id"`
}
