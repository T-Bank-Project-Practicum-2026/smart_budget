package kafka

import (
	"errors"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewClient(brokers []string, opts ...kgo.Opt) (*kgo.Client, error) {
	if len(brokers) == 0 {
		return nil, errors.New("kafka: at least one broker is required")
	}

	options := []kgo.Opt{kgo.SeedBrokers(brokers...)}
	options = append(options, opts...)

	return kgo.NewClient(options...)
}
