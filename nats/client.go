package natsclient

import (
	"encoding/json"
	"fmt"
	"time"

	nats "github.com/nats-io/nats.go"
)

//NatsClient is the state of the client
type NatsClient struct {
	client        *nats.Conn
	subscriptions map[string]*nats.Subscription
}

//NewNatsClient creates a new NATS client and attempts to connect to the specified endpoint
func NewNatsClient(endpoint string) (*NatsClient, error) {
	client, err := nats.Connect(endpoint)

	if err != nil {
		return nil, err
	}

	return &NatsClient{
		client:        client,
		subscriptions: make(map[string]*nats.Subscription),
	}, nil
}

//Publish sends a message to subject subscriptions
func (c *NatsClient) Publish(subject string, msg interface{}) error {
	bytes, err := json.Marshal(msg)

	if err != nil {
		return fmt.Errorf("Failed to marshal message: %s", err)
	}

	err = c.client.Publish(subject, bytes)

	if err != nil {
		return fmt.Errorf("Failed to publish message: %s", err)
	}

	return nil
}

//Request sends a message to subject subscriptions
func (c *NatsClient) Request(subject string, msg interface{}) (*nats.Msg, error) {
	bytes, err := json.Marshal(msg)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal Request payload: %s", err)
	}

	reqMsg, err := c.client.Request(subject, bytes, 5*time.Second)

	if err != nil {
		return nil, fmt.Errorf("Failed to send Request: %s", err)
	}

	return reqMsg, nil
}

//Subscribe to a nats subject with a callback
func (c *NatsClient) Subscribe(subject string, callback func(msg *nats.Msg)) error {
	if err := c.validateSubscription(subject); err != nil {
		return err
	}

	sub, err := c.client.Subscribe(subject, callback)

	if err != nil {
		return err
	}

	c.subscriptions[subject] = sub

	return nil
}

//QueueSubscribe to a nats subject and queue group with a callback
func (c *NatsClient) QueueSubscribe(subject string, queue string, callback func(msg *nats.Msg)) error {
	if err := c.validateSubscription(subject); err != nil {
		return err
	}

	sub, err := c.client.QueueSubscribe(subject, queue, callback)

	if err != nil {
		return err
	}

	c.subscriptions[subject] = sub

	return nil
}

//ChanSubscribe to a nats subject with a channel
func (c *NatsClient) ChanSubscribe(subject string) (<-chan *nats.Msg, error) {
	if err := c.validateSubscription(subject); err != nil {
		return nil, err
	}

	channel := make(chan *nats.Msg, 1)

	sub, err := c.client.ChanSubscribe(subject, channel)

	if err != nil {
		return nil, err
	}

	c.subscriptions[subject] = sub

	return channel, nil
}

//ChanQueueSubscribe to a nats subject and queue group with a channel
func (c *NatsClient) ChanQueueSubscribe(subject string, queue string) (<-chan *nats.Msg, error) {
	if err := c.validateSubscription(subject); err != nil {
		return nil, err
	}

	channel := make(chan *nats.Msg, 1)

	sub, err := c.client.QueueSubscribeSyncWithChan(subject, queue, channel)

	if err != nil {
		return nil, err
	}

	c.subscriptions[subject] = sub

	return channel, nil
}

//Unsubscribe drains a subscription
func (c *NatsClient) Unsubscribe(subject string) error {
	sub, exists := c.subscriptions[subject]

	if !exists {
		return nil
	}

	err := sub.Drain()

	if err != nil {
		return err
	}

	delete(c.subscriptions, subject)

	return nil
}

//Shutdown gracefull cleans up subscriptions and the client
func (c *NatsClient) Shutdown() {
	for _, sub := range c.subscriptions {
		sub.Drain()
	}

	c.client.Drain()
}

func (c *NatsClient) validateSubscription(subject string) error {
	if _, exists := c.subscriptions[subject]; exists {
		return fmt.Errorf("Subscription for '%s' already exists", subject)
	}

	return nil
}
