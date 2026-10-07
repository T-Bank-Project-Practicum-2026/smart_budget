package push

import (
	"encoding/json"
	"errors"
	"time"
)

const Topic = "notifications.requested.v1"

type Event struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

func Decode(data []byte) (Event, error) {
	var event Event

	if len(data) > 8192 || json.Unmarshal(data, &event) != nil || event.ID == "" || len(event.ID) > 100 || event.UserID == "" || len(event.UserID) > 100 || event.Title == "" || len(event.Title) > 200 || len(event.Body) > 1000 || event.CreatedAt.IsZero() {
		return event, errors.New("invalid push event")
	}

	return event, nil
}
