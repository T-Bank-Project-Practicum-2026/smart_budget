package push

import (
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventWireAndStoredRoundTrip(t *testing.T) {
	event := Event{ID: "event", UserID: "user", Title: "Бюджет", Body: "Лимит достигнут", CreatedAt: time.Date(2026, 10, 22, 12, 0, 0, 123456789, time.UTC)}
	stored, err := Encode(event)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decode(stored)
	if err != nil || decoded != event {
		t.Fatalf("stored event: %+v %v", decoded, err)
	}

	if _, err := Decode(append([]byte(`{"unexpected":true,`), stored[1:]...)); err == nil {
		t.Fatal("accepted unknown ProtoJSON field")
	}

	wire, err := Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	message := &servicesv1.NotificationRequested{}
	if err := proto.Unmarshal(wire, message); err != nil || message.UserId != event.UserID || !message.CreatedAt.AsTime().Equal(event.CreatedAt) {
		t.Fatalf("generated contract: %v %v", message, err)
	}

	// New fields from a newer producer must not break existing consumers.
	wire = protowire.AppendTag(wire, 100, protowire.VarintType)
	wire = protowire.AppendVarint(wire, 1)
	decoded, err = Unmarshal(wire)
	if err != nil || decoded != event {
		t.Fatalf("wire event: %+v %v", decoded, err)
	}

	if _, err := Unmarshal(stored); err == nil {
		t.Fatal("JSON accepted as binary Protobuf")
	}
}

func TestRejectInvalidWireEvents(t *testing.T) {
	for _, scenario := range []string{"missing_time", "invalid_time", "zero_time", "empty_id", "long_title", "long_body"} {
		t.Run(scenario, func(t *testing.T) {
			message := &servicesv1.NotificationRequested{Id: "event", UserId: "user", Title: "title", CreatedAt: timestamppb.New(time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC))}

			switch scenario {
			case "missing_time":
				message.CreatedAt = nil

			case "invalid_time":
				message.CreatedAt.Nanos = -1

			case "zero_time":
				message.CreatedAt = timestamppb.New(time.Time{})

			case "empty_id":
				message.Id = ""

			case "long_title":
				message.Title = strings.Repeat("x", 201)

			case "long_body":
				message.Body = strings.Repeat("x", 1001)
			}

			wire, err := proto.Marshal(message)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := Unmarshal(wire); err == nil {
				t.Fatal("accepted invalid event")
			}
		})
	}

	for _, wire := range [][]byte{nil, {0xff}, []byte(strings.Repeat("x", 8193))} {
		if _, err := Unmarshal(wire); err == nil {
			t.Fatal("accepted invalid wire data")
		}
	}
}
