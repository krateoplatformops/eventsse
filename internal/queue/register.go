package queue

import (
	"fmt"
	"net/url"
)

var (
	register = make(map[string]BrokerBuilder, 0)
)

// BrokerBuilder instantiates a new Broker based on the given uri.
type BrokerBuilder func(uri string) (Broker, error)

// Register registers a new BrokerBuilder to be used by NewBroker, this function
// should be used in an init function in the implementation packages such as
// `amqp` and `memory`.
func Register(name string, b BrokerBuilder) {
	register[name] = b
}

// NewBroker creates a new Broker based on the given URI. In order to register
// different implementations the package should be imported.
func NewBroker(uri string) (Broker, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("malformed uri '%s': %w", uri, err)
	}

	if url.Scheme == "" {
		return nil, fmt.Errorf("malformed uri '%s'", uri)
	}

	b, ok := register[url.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported protocol '%s': %w", uri, err)
	}

	return b(uri)
}
