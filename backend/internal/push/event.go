package push

import (
	"errors"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const Topic = "notifications.requested.v2"

type Event struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
}

// Encode готовит ProtoJSON для хранения в outbox.
func Encode(event Event) ([]byte, error) {
	message := event.message()
	if err := validate(message); err != nil {
		return nil, err
	}

	return protojson.Marshal(message)
}

func Decode(data []byte) (Event, error) {
	message := &servicesv1.NotificationRequested{}
	if len(data) > 8192 {
		return Event{}, errors.New("invalid push event")
	}

	if err := protojson.Unmarshal(data, message); err != nil {
		return Event{}, errors.New("invalid push event")
	}

	return fromMessage(message)
}

// Marshal готовит бинарное Protobuf-сообщение для Kafka.
func Marshal(event Event) ([]byte, error) {
	message := event.message()
	if err := validate(message); err != nil {
		return nil, err
	}

	return proto.Marshal(message)
}

func Unmarshal(data []byte) (Event, error) {
	message := &servicesv1.NotificationRequested{}
	if len(data) > 8192 || proto.Unmarshal(data, message) != nil {
		return Event{}, errors.New("invalid push event")
	}

	return fromMessage(message)
}

func (event Event) message() *servicesv1.NotificationRequested {
	return &servicesv1.NotificationRequested{Id: event.ID, UserId: event.UserID, Title: event.Title, Body: event.Body, CreatedAt: timestamppb.New(event.CreatedAt)}
}

func fromMessage(message *servicesv1.NotificationRequested) (Event, error) {
	if err := validate(message); err != nil {
		return Event{}, err
	}

	return Event{ID: message.Id, UserID: message.UserId, Title: message.Title, Body: message.Body, CreatedAt: message.CreatedAt.AsTime()}, nil
}

func validate(message *servicesv1.NotificationRequested) error {
	if message.Id == "" || len(message.Id) > 100 || message.UserId == "" || len(message.UserId) > 100 || message.Title == "" || len(message.Title) > 200 || len(message.Body) > 1000 || message.CreatedAt.CheckValid() != nil || message.CreatedAt.AsTime().IsZero() {
		return errors.New("invalid push event")
	}

	return nil
}
