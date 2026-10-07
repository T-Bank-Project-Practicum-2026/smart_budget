package outbox

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"os"
	"smart_budget/internal/kafka"
	"smart_budget/internal/notifications"
	"smart_budget/internal/push"
	"smart_budget/internal/testutil"
	"strings"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

func TestQuarantinePublishReplayAndCleanup(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("run make test for Kafka integration")
	}

	db := testutil.Database(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)

	defer cancel()

	topic := "budget-test-" + strings.ToLower(rand.Text())
	client, err := kafka.NewClient(strings.Split(brokers, ","))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	create := kmsg.NewPtrCreateTopicsRequest()
	create.Topics = []kmsg.CreateTopicsRequestTopic{{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}}
	response, err := create.RequestWith(ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	for _, result := range response.Topics {
		if result.ErrorCode != 0 {
			t.Fatalf("create topic: %d", result.ErrorCode)
		}
	}

	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)

		defer stop()

		request := kmsg.NewPtrDeleteTopicsRequest()
		request.TopicNames = []string{topic}
		if _, err := request.RequestWith(cleanup, client); err != nil {
			t.Error(err)
		}
	}()

	consumer, err := kafka.NewClient(strings.Split(brokers, ","), kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{topic: {0: kgo.NewOffset().AtStart()}}))
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()

	_, err = db.Exec(ctx, `INSERT INTO app.users(id,email,password_hash) VALUES('u','u@example.test','unused');
 INSERT INTO app.notifications(id,user_id,request_key,title,body) VALUES('bad','u','bad','t','b'),('good','u','good','t','b');
 INSERT INTO app.outbox(id,payload,created_at) VALUES('bad','{}',now()-interval '1 minute');`)
	if err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(push.Event{ID: "good", UserID: "u", Title: "test", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec(ctx, "INSERT INTO app.outbox(id,payload) VALUES('good',$1)", payload); err != nil {
		t.Fatal(err)
	}

	publisher := &Publisher{DB: db, Kafka: client, Topic: topic}

	for range 2 {
		sent, err := publisher.PublishOne(ctx)
		if err != nil || !sent {
			t.Fatalf("publish: sent=%v err=%v", sent, err)
		}
	}

	var quarantined bool

	if err = db.QueryRow(ctx, "SELECT failed_at IS NOT NULL AND last_error='invalid_payload' FROM app.outbox WHERE id='bad'").Scan(&quarantined); err != nil || !quarantined {
		t.Fatalf("quarantine=%v err=%v", quarantined, err)
	}

	// Simulate a crash after broker acknowledgement but before the DB mark.
	if _, err = db.Exec(ctx, "UPDATE app.outbox SET published_at=NULL WHERE id='good'"); err != nil {
		t.Fatal(err)
	}

	if sent, err := publisher.PublishOne(ctx); err != nil || !sent {
		t.Fatalf("republish: %v %v", sent, err)
	}

	svc := &notifications.Service{DB: db}
	count := 0

	for count < 2 {
		fetches := consumer.PollRecords(ctx, 2)
		if errs := fetches.Errors(); len(errs) > 0 {
			t.Fatal(errs)
		}

		for _, record := range fetches.Records() {
			if err := svc.Accept(ctx, record); err != nil {
				t.Fatal(err)
			}

			count++
		}
	}

	var inbox int

	if err = db.QueryRow(ctx, "SELECT count(*) FROM notifications.inbox WHERE id='good'").Scan(&inbox); err != nil || inbox != 1 {
		t.Fatalf("inbox=%d err=%v", inbox, err)
	}

	if _, err = db.Exec(ctx, "UPDATE app.outbox SET published_at=now()-interval '31 days' WHERE id='good'"); err != nil {
		t.Fatal(err)
	}

	if err = publisher.Cleanup(ctx); err != nil {
		t.Fatal(err)
	}

	var retained int

	if err = db.QueryRow(ctx, "SELECT count(*) FROM app.outbox").Scan(&retained); err != nil || retained != 1 {
		t.Fatalf("retain quarantine, remove published: %d %v", retained, err)
	}
}
